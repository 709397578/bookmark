package snapshot

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// SnapshotService 快照服务
type SnapshotService struct {
	uploadDir  string
	chromePath string
}

// NewSnapshotService 创建快照服务
// chromePath 可选，指定 Chrome/Chromium 可执行文件路径，为空则自动查找
func NewSnapshotService(uploadDir string, chromePath string) *SnapshotService {
	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("无法创建上传目录: %v", err))
	}

	if chromePath == "" {
		chromePath = os.Getenv("CHROME_PATH")
	}
	// 验证路径是否存在，不存在则清空，让 chromedp 自动在 PATH 中查找
	if chromePath != "" {
		if _, err := os.Stat(chromePath); os.IsNotExist(err) {
			chromePath = ""
		}
	}

	return &SnapshotService{
		uploadDir:  uploadDir,
		chromePath: chromePath,
	}
}

// ValidateURL 验证URL是否安全（防SSRF）
func ValidateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL格式无效: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("不支持的URL协议: %s", parsed.Scheme)
	}

	host := parsed.Hostname()

	// 检查是否为内部主机名
	privateHosts := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0",
		"metadata.google.internal", "metadata.tencentyun.com",
		"100.100.100.200"}
	for _, h := range privateHosts {
		if strings.EqualFold(host, h) {
			return fmt.Errorf("禁止访问内部地址: %s", host)
		}
	}

	// 检查是否为内网IP段
	ips, err := net.LookupIP(host)
	if err != nil {
		// DNS解析失败，记录但继续（可能是有效的外部域名）
		return nil
	}

	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
			return fmt.Errorf("禁止访问内网地址: %s (%s)", host, ip.String())
		}
	}

	return nil
}

// GenerateSnapshot 生成网页快照（PDF格式）
func (s *SnapshotService) GenerateSnapshot(ctx context.Context, rawURL string) (string, error) {
	// 验证URL安全性
	if err := ValidateURL(rawURL); err != nil {
		return "", fmt.Errorf("URL验证失败: %w", err)
	}

	// 生成唯一的文件名
	filename := fmt.Sprintf("snapshot_%d.pdf", time.Now().UnixNano())
	filepath := filepath.Join(s.uploadDir, filename)

	// 清除代理环境变量，避免 Chrome 被系统代理拦截
	proxyVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "no_proxy", "ALL_PROXY", "all_proxy"}
	saved := make(map[string]string)
	for _, v := range proxyVars {
		saved[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	defer func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	}()

	// 配置Chrome选项
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("proxy-server", "direct://"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
	}
	if s.chromePath != "" {
		opts = append(opts, chromedp.ExecPath(s.chromePath))
	}
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	taskCtx, cancel = context.WithTimeout(taskCtx, 30*time.Second)
	defer cancel()

	// 存储PDF内容
	var buf []byte
	var pageTitle string
	var docURL string

	// 移除弹窗和遮罩层，展开滚动区域
	pageJS := `
		(() => {
			document.documentElement.style.overflow = '';
			document.body.style.overflow = '';
			// 1. 移除各种弹窗/遮罩
			document.querySelectorAll('div').forEach(el => {
				const style = window.getComputedStyle(el);
				if (style.position === 'fixed' || style.position === 'absolute') {
					const cls = el.className.toLowerCase();
					if (cls.includes('modal') || cls.includes('overlay') || cls.includes('mask') ||
						cls.includes('backdrop') || cls.includes('dialog') || cls.includes('popup') ||
						cls.includes('signflow') || cls.includes('sign') || cls.includes('login') ||
						cls.includes('guide') || cls.includes('tips') || cls.includes('toast') ||
						cls.includes('notice') || cls.includes('float') ||
						el.getAttribute('aria-modal') === 'true' ||
						el.getAttribute('role') === 'dialog') {
						el.remove();
						return;
					}
					if (el.clientWidth >= window.innerWidth * 0.9 && el.clientHeight >= window.innerHeight * 0.9) {
						el.remove();
					}
				}
				if (style.position === 'sticky' && parseInt(style.zIndex) > 100) {
					el.remove();
				}
			});
			// 2. 展开所有有滚动条的区域（代码块等），让完整内容可见
			document.querySelectorAll('pre, code, .highlight, .code-block, [class*="code"]').forEach(el => {
				el.style.overflow = 'visible';
				el.style.maxHeight = 'none';
			});
			// 3. 展开所有带 overflow scroll/auto 的容器
			document.querySelectorAll('div, section, article, main, aside').forEach(el => {
				const style = window.getComputedStyle(el);
				const ov = style.overflow + style.overflowX + style.overflowY;
				if (ov.includes('scroll') || ov.includes('auto')) {
					el.style.overflow = 'visible';
					el.style.maxHeight = 'none';
				}
			});
		})();
	`

	// 执行chromedp任务
	err := chromedp.Run(taskCtx,
		// 导航到URL
		chromedp.Navigate(rawURL),
		// 等待页面加载完成
		chromedp.WaitReady("body"),
		// 等待一段时间，让JavaScript执行
		chromedp.Sleep(3*time.Second),
		// 移除弹窗元素并展开滚动区域
		chromedp.EvaluateAsDevTools(pageJS, nil),
		// 额外等待弹窗动画完成
		chromedp.Sleep(1*time.Second),
		// 获取当前页面URL（诊断用）
		chromedp.EvaluateAsDevTools(`document.location.href`, &docURL),
		// 获取页面标题
		chromedp.Title(&pageTitle),
		// 打印页面为PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	)

	if err != nil {
		return "", fmt.Errorf("生成快照失败: %w", err)
	}

	// 诊断日志
	log.Printf("[快照] 目标URL: %s, 实际访问: %s, 页面标题: %s", rawURL, docURL, pageTitle)
	if docURL != "" && urlHost(docURL) != urlHost(rawURL) {
		log.Printf("[快照] 警告: 实际访问域名 %s 与目标域名 %s 不一致", urlHost(docURL), urlHost(rawURL))
	}

	// 创建文件
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入PDF内容
	if _, err := file.Write(buf); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 返回相对路径，前端通过 nginx 反代访问
	return "/snapshots/" + filename, nil
}

func urlHost(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}



// DeleteSnapshot 删除快照文件
func (s *SnapshotService) DeleteSnapshot(snapshotURL string) error {
	// 从URL中提取文件名
	filename := filepath.Base(snapshotURL)
	filepath := filepath.Join(s.uploadDir, filename)

	// 删除文件
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除快照文件失败: %w", err)
	}

	return nil
}
