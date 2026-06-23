package handlers

import (
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"pintree-backend/config"
	"pintree-backend/internal"
	"pintree-backend/models"
	"pintree-backend/pkg/snapshot"
	"pintree-backend/utils"
	"github.com/gin-gonic/gin"
)

// BookmarkHandler 书签处理器
type BookmarkHandler struct {
	snapshotService *snapshot.SnapshotService
}

func NewBookmarkHandler(cfg *config.Config) *BookmarkHandler {
	return &BookmarkHandler{
		snapshotService: snapshot.NewSnapshotService(
			cfg.SnapshotDir,
			cfg.ChromePath,
		),
	}
}

// GetBookmarks 获取书签列表（分页）
func (h *BookmarkHandler) GetBookmarks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	collectionID := c.Query("collectionId")
	folderID := c.Query("folderId")
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}
	
	offset := (page - 1) * pageSize
	
	var bookmarks []models.Bookmark
	var total int64
	
	query := internal.DB.Preload("Collection").Preload("Folder").Preload("Tags")
	
	// 如果指定了收藏夹ID，检查该收藏夹是否公开
	if collectionID != "" {
		var collection models.Collection
		if err := internal.DB.First(&collection, "id = ?", collectionID).Error; err == nil {
			userID := c.GetString("userID")
			
			// 如果收藏夹不公开且用户未登录或不是所有者，返回空列表
			if !collection.IsPublic && (userID == "" || userID != collection.UserID) {
				response := utils.PaginatedResponse{
					Data:        []models.Bookmark{},
					Total:       0,
					CurrentPage: page,
					TotalPages:  0,
					PageSize:    pageSize,
				}
				c.JSON(http.StatusOK, utils.SuccessResponse(response, ""))
				return
			}
		}
		
		query = query.Where("collection_id = ?", collectionID)
	}
	
	if folderID != "" && folderID != "none" {
		query = query.Where("folder_id = ?", folderID)
	} else if folderID == "none" {
		query = query.Where("folder_id IS NULL")
	}
	
	// 获取总数
	query.Model(&models.Bookmark{}).Count(&total)
	
	// 获取分页数据
	if err := query.Order("sort_order ASC, updated_at DESC").Offset(offset).Limit(pageSize).Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取书签失败"))
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	response := utils.PaginatedResponse{
		Data:        bookmarks,
		Total:       total,
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    pageSize,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(response, ""))
}

// GetBookmarkByID 根据ID获取书签
func (h *BookmarkHandler) GetBookmarkByID(c *gin.Context) {
	id := c.Param("id")
	
	var bookmark models.Bookmark
	if err := internal.DB.Preload("Collection").Preload("Folder").Preload("Tags").First(&bookmark, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("书签不存在"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(bookmark, ""))
}

// CreateBookmark 创建书签
func (h *BookmarkHandler) CreateBookmark(c *gin.Context) {
	var req struct {
		Title        string   `json:"title" binding:"required"`
		URL          string   `json:"url" binding:"required"`
		Description  string   `json:"description"`
		Icon         string   `json:"icon"`
		CollectionID string   `json:"collectionId" binding:"required"`
		FolderID     *string  `json:"folderId"`
		Tags         []string `json:"tags"`
		IsFeatured   bool     `json:"isFeatured"`
		SortOrder    int      `json:"sortOrder"`
		HasSnapshot  bool     `json:"hasSnapshot"`
		SnapshotURL  string   `json:"snapshotUrl"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误：标题、URL和收藏夹为必填项"))
		return
	}

	// 验证文件夹是否存在
	if req.FolderID != nil && *req.FolderID != "none" {
		var folder models.Folder
		if err := internal.DB.Where("id = ? AND collection_id = ?", *req.FolderID, req.CollectionID).First(&folder).Error; err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse("指定的文件夹不存在或不属于该收藏夹"))
			return
		}
	}

	bookmark := models.Bookmark{
		Title:        req.Title,
		URL:          req.URL,
		Description:  req.Description,
		Icon:         req.Icon,
		CollectionID: req.CollectionID,
		FolderID:     req.FolderID,
		IsFeatured:   req.IsFeatured,
		SortOrder:    req.SortOrder,
	}

	if err := internal.DB.Create(&bookmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("创建书签失败，请检查所有字段是否正确"))
		return
	}

	// 后台异步生成网页快照
	go func(bmID, bmURL string) {
		snapshotURL, err := h.snapshotService.GenerateSnapshot(context.Background(), bmURL)
		if err != nil {
			log.Printf("后台生成快照失败 [书签ID: %s]: %v", bmID, err)
			return
		}
		if err := internal.DB.Model(&models.Bookmark{}).Where("id = ?", bmID).Updates(map[string]interface{}{
			"has_snapshot": true,
			"snapshot_url": snapshotURL,
		}).Error; err != nil {
			log.Printf("更新书签快照失败 [书签ID: %s]: %v", bmID, err)
		} else {
			log.Printf("后台生成快照成功 [书签ID: %s]: %s", bmID, snapshotURL)
		}
	}(bookmark.ID, bookmark.URL)

	// 重新加载关联数据
	internal.DB.Preload("Collection").Preload("Folder").Preload("Tags").First(&bookmark, bookmark.ID)

	c.JSON(http.StatusCreated, utils.SuccessResponse(bookmark, "创建成功"))
}

// UpdateBookmark 更新书签
func (h *BookmarkHandler) UpdateBookmark(c *gin.Context) {
	id := c.Param("id")
	
	var bookmark models.Bookmark
	if err := internal.DB.First(&bookmark, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("书签不存在"))
		return
	}

	var req struct {
		Title       string  `json:"title"`
		URL         string  `json:"url"`
		Description string  `json:"description"`
		Icon        string  `json:"icon"`
		FolderID    *string `json:"folderId"`
		IsFeatured  bool    `json:"isFeatured"`
		SortOrder   int     `json:"sortOrder"`
		HasSnapshot bool    `json:"hasSnapshot"`
		SnapshotURL string  `json:"snapshotUrl"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}

	updates := map[string]interface{}{
		"title":        req.Title,
		"url":          req.URL,
		"description":  req.Description,
		"icon":         req.Icon,
		"is_featured":  req.IsFeatured,
		"sort_order":   req.SortOrder,
		"has_snapshot": req.HasSnapshot,
		"snapshot_url": req.SnapshotURL,
	}

	if req.FolderID != nil {
		if *req.FolderID != "none" {
			updates["folder_id"] = req.FolderID
		} else {
			updates["folder_id"] = nil
		}
	}

	// 处理快照更新
	if req.HasSnapshot {
		// 如果之前没有快照，生成新快照
		if !bookmark.HasSnapshot {
			snapshotURL, err := h.snapshotService.GenerateSnapshot(context.Background(), req.URL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse("生成快照失败: "+err.Error()))
				return
			}
			updates["has_snapshot"] = true
			updates["snapshot_url"] = snapshotURL
		}
		// 如果URL发生变化，删除旧快照并生成新快照
		if bookmark.URL != req.URL && bookmark.SnapshotURL != "" {
			// 删除旧快照
			if err := h.snapshotService.DeleteSnapshot(bookmark.SnapshotURL); err != nil {
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse("删除旧快照失败: "+err.Error()))
				return
			}
			// 生成新快照
			snapshotURL, err := h.snapshotService.GenerateSnapshot(context.Background(), req.URL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse("生成快照失败: "+err.Error()))
				return
			}
			updates["snapshot_url"] = snapshotURL
		}
	} else {
		// 如果禁用快照，删除现有快照
		if bookmark.HasSnapshot && bookmark.SnapshotURL != "" {
			if err := h.snapshotService.DeleteSnapshot(bookmark.SnapshotURL); err != nil {
				c.JSON(http.StatusInternalServerError, utils.ErrorResponse("删除快照失败: "+err.Error()))
				return
			}
			updates["has_snapshot"] = false
			updates["snapshot_url"] = ""
		}
	}

	if err := internal.DB.Model(&bookmark).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新失败"))
		return
	}

	// 重新加载数据
	internal.DB.Preload("Collection").Preload("Folder").Preload("Tags").First(&bookmark, id)

	c.JSON(http.StatusOK, utils.SuccessResponse(bookmark, "更新成功"))
}

// DeleteBookmark 删除书签
func (h *BookmarkHandler) DeleteBookmark(c *gin.Context) {
	id := c.Param("id")
	
	var bookmark models.Bookmark
	if err := internal.DB.First(&bookmark, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("书签不存在"))
		return
	}

	// 删除关联的快照文件
	if bookmark.SnapshotURL != "" {
		if err := h.snapshotService.DeleteSnapshot(bookmark.SnapshotURL); err != nil {
			log.Printf("删除快照文件失败 [书签ID: %s]: %v", id, err)
		}
	}

	if err := internal.DB.Delete(&bookmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("删除失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "删除成功"))
}

// GenerateSnapshot 生成书签快照（异步）
func (h *BookmarkHandler) GenerateSnapshot(c *gin.Context) {
	id := c.Param("id")

	var bookmark models.Bookmark
	if err := internal.DB.First(&bookmark, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("书签不存在"))
		return
	}

	log.Printf("开始异步生成快照，书签ID: %s, URL: %s", id, bookmark.URL)

	// 先清理旧的快照信息，让前端立即看到"生成中"状态
	internal.DB.Model(&bookmark).Updates(map[string]interface{}{
		"has_snapshot": false,
		"snapshot_url": "",
	})

	c.JSON(http.StatusOK, utils.SuccessResponse(gin.H{
		"id":          id,
		"hasSnapshot": false,
		"message":     "快照正在后台生成，请稍后刷新查看",
	}, "正在生成快照"))

	// 异步生成快照
	go func(bmID, bmURL string) {
		snapshotURL, err := h.snapshotService.GenerateSnapshot(context.Background(), bmURL)
		if err != nil {
			log.Printf("异步生成快照失败 [书签ID: %s]: %v", bmID, err)
			return
		}
		if err := internal.DB.Model(&models.Bookmark{}).Where("id = ?", bmID).Updates(map[string]interface{}{
			"has_snapshot": true,
			"snapshot_url": snapshotURL,
		}).Error; err != nil {
			log.Printf("异步更新书签快照失败 [书签ID: %s]: %v", bmID, err)
		} else {
			log.Printf("异步生成快照成功 [书签ID: %s]: %s", bmID, snapshotURL)
		}
	}(bookmark.ID, bookmark.URL)
}

// SearchBookmarks 搜索书签
func (h *BookmarkHandler) SearchBookmarks(c *gin.Context) {
	query := c.Query("q")
	
	if query == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("搜索关键词不能为空"))
		return
	}

	var bookmarks []models.Bookmark
	searchTerm := "%" + query + "%"
	
	if err := internal.DB.Preload("Collection").Preload("Folder").Preload("Tags").
		Where("title LIKE ? OR url LIKE ? OR description LIKE ?", searchTerm, searchTerm, searchTerm).
		Order("updated_at DESC").
		Limit(50).
		Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("搜索失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(bookmarks, ""))
}

// ImportBookmarks 导入书签
func (h *BookmarkHandler) ImportBookmarks(c *gin.Context) {
	// 获取请求体中的JSON数据
	var data []interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请提供有效的JSON数据"))
		return
	}

	// 获取用户ID
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("未登录或登录已过期"))
		return
	}

	// 获取目标收藏夹ID
	collectionID := c.Query("collectionId")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请指定目标收藏夹"))
		return
	}

	// 验证收藏夹是否存在且属于当前用户
	var collection models.Collection
	if err := internal.DB.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("指定的收藏夹不存在或不属于您"))
		return
	}

	// 用于统计导入结果
	importedCount := 0;
	errorCount := 0
	var createdFolders []models.Folder
	var createdBookmarks []models.Bookmark

	// 递归处理书签数据
	// 使用函数变量解决递归调用问题
	var processItems func([]interface{}, *string)
	processItems = func(items []interface{}, parentID *string) {
		for _, item := range items {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				errorCount++
				continue
			}

			// 获取类型
			itemType, ok := itemMap["type"].(string)
			if !ok {
				errorCount++
				continue
			}

			if itemType == "folder" {
				// 创建文件夹
				folderName, _ := itemMap["title"].(string)
				if folderName == "" {
					folderName = "未命名文件夹"
				}

				// 忽略icon字段，因为系统有处理icon的函数
				newFolder := models.Folder{
					Name:        folderName,
					CollectionID: collectionID,
					ParentID:    parentID,
				}

				if err := internal.DB.Create(&newFolder).Error; err != nil {
					fmt.Printf("创建文件夹失败: %v\n", err)
					errorCount++
					continue
				}

				createdFolders = append(createdFolders, newFolder)
				importedCount++

				// 递归处理子项
				if children, ok := itemMap["children"].([]interface{}); ok {
					processItems(children, &newFolder.ID)
				}
			} else if itemType == "link" {
				// 创建书签
				title, _ := itemMap["title"].(string)
				url, _ := itemMap["url"].(string)
				if title == "" || url == "" {
					errorCount++
					continue
				}

				// 验证URL格式
				if !isValidURL(url) {
					errorCount++
					continue
				}

				// 忽略icon字段，因为系统有处理icon的函数
				newBookmark := models.Bookmark{
					Title:        title,
					URL:          url,
					Description:  getString(itemMap, "description"),
					CollectionID: collectionID,
					FolderID:     parentID,
					SortOrder:    getInt(itemMap, "sortOrder"),
				}

				if err := internal.DB.Create(&newBookmark).Error; err != nil {
					fmt.Printf("创建书签失败: %v\n", err)
					errorCount++
					continue
				}

				createdBookmarks = append(createdBookmarks, newBookmark)
				importedCount++
			}
		}
	}

	// 开始处理
	processItems(data, nil)

	result := map[string]interface{}{
		"importedCount": importedCount,
		"errorCount":   errorCount,
		"folders":      createdFolders,
		"bookmarks":    createdBookmarks,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(result, "导入成功"))
}

// ExportBookmarks 导出收藏夹书签为JSON格式（与导入格式一致）
func (h *BookmarkHandler) ExportBookmarks(c *gin.Context) {
	collectionID := c.Query("collectionId")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请指定收藏夹"))
		return
	}

	// 获取所有文件夹
	var folders []models.Folder
	if err := internal.DB.Where("collection_id = ?", collectionID).Order("sort_order ASC").Find(&folders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取文件夹失败"))
		return
	}

	// 获取所有书签
	var bookmarks []models.Bookmark
	if err := internal.DB.Where("collection_id = ?", collectionID).Order("sort_order ASC, updated_at DESC").Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取书签失败"))
		return
	}

	// 构建文件夹树
	folderMap := make(map[string]map[string]interface{})
	var rootFolders []map[string]interface{}

	// 先创建所有文件夹节点
	for _, folder := range folders {
		folderNode := map[string]interface{}{
			"type":  "folder",
			"title": folder.Name,
		}
		folderMap[folder.ID] = folderNode
	}

	// 建立父子关系
	for _, folder := range folders {
		folderNode := folderMap[folder.ID]
		if folder.ParentID != nil && *folder.ParentID != "" {
			if parent, ok := folderMap[*folder.ParentID]; ok {
				if children, exists := parent["children"]; exists {
					parent["children"] = append(children.([]interface{}), folderNode)
				} else {
					parent["children"] = []interface{}{folderNode}
				}
			} else {
				rootFolders = append(rootFolders, folderNode)
			}
		} else {
			rootFolders = append(rootFolders, folderNode)
		}
	}

	// 将书签添加到对应的文件夹或根目录
	for _, bookmark := range bookmarks {
		bookmarkNode := map[string]interface{}{
			"type":       "link",
			"title":      bookmark.Title,
			"url":        bookmark.URL,
			"description": bookmark.Description,
			"sortOrder":  bookmark.SortOrder,
		}

		if bookmark.FolderID != nil && *bookmark.FolderID != "" {
			if folder, ok := folderMap[*bookmark.FolderID]; ok {
				if children, exists := folder["children"]; exists {
					folder["children"] = append(children.([]interface{}), bookmarkNode)
				} else {
					folder["children"] = []interface{}{bookmarkNode}
				}
			} else {
				rootFolders = append(rootFolders, bookmarkNode)
			}
		} else {
			rootFolders = append(rootFolders, bookmarkNode)
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(rootFolders, "导出成功"))
}

// 辅助函数：从map中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：从map中获取整数值
func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
	}
	return 0
}

// 辅助函数：验证URL格式
func isValidURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

// ExportBookmarksHTML 导出书签为 Netscape/Chrome HTML 格式
func (h *BookmarkHandler) ExportBookmarksHTML(c *gin.Context) {
	collectionID := c.Query("collectionId")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请指定收藏夹"))
		return
	}

	var collection models.Collection
	if err := internal.DB.First(&collection, "id = ?", collectionID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("收藏夹不存在"))
		return
	}

	var folders []models.Folder
	if err := internal.DB.Where("collection_id = ?", collectionID).Order("sort_order ASC").Find(&folders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取文件夹失败"))
		return
	}

	var bookmarks []models.Bookmark
	if err := internal.DB.Where("collection_id = ?", collectionID).Order("sort_order ASC, updated_at DESC").Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取书签失败"))
		return
	}

	now := strconv.FormatInt(time.Now().Unix(), 10)

	// 构建文件夹ID到子文件夹的映射
	folderChildren := make(map[string][]models.Folder)
	folderBookmarkMap := make(map[string][]models.Bookmark)
	var rootFolders []models.Folder
	var rootBookmarks []models.Bookmark

	for _, folder := range folders {
		if folder.ParentID != nil && *folder.ParentID != "" {
			folderChildren[*folder.ParentID] = append(folderChildren[*folder.ParentID], folder)
		} else {
			rootFolders = append(rootFolders, folder)
		}
	}

	for _, bookmark := range bookmarks {
		if bookmark.FolderID != nil && *bookmark.FolderID != "" {
			folderBookmarkMap[*bookmark.FolderID] = append(folderBookmarkMap[*bookmark.FolderID], bookmark)
		} else {
			rootBookmarks = append(rootBookmarks, bookmark)
		}
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE NETSCAPE-Bookmark-file-1>\n")
	sb.WriteString("<!-- This is an automatically generated file.\n")
	sb.WriteString("     It will be read and overwritten.\n")
	sb.WriteString("     DO NOT EDIT! -->\n")
	sb.WriteString("<META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html; charset=UTF-8\">\n")
	sb.WriteString("<TITLE>Bookmarks</TITLE>\n")
	sb.WriteString("<H1>Bookmarks</H1>\n")
	sb.WriteString("<DL><p>\n")

	var writeFolder func(folder models.Folder, depth int)
	writeFolder = func(folder models.Folder, depth int) {
		indent := strings.Repeat("    ", depth)
		sb.WriteString(fmt.Sprintf("%s<DT><H3 ADD_DATE=\"%s\">%s</H3>\n", indent, now, html.EscapeString(folder.Name)))
		sb.WriteString(fmt.Sprintf("%s<DL><p>\n", indent))

		// 写入此文件夹下的书签
		for _, bm := range folderBookmarkMap[folder.ID] {
			sb.WriteString(fmt.Sprintf("%s    <DT><A HREF=\"%s\" ADD_DATE=\"%s\">%s</A>\n",
				indent, html.EscapeString(bm.URL),
				strconv.FormatInt(bm.CreatedAt.Unix(), 10),
				html.EscapeString(bm.Title)))
		}

		// 写入子文件夹
		for _, childFolder := range folderChildren[folder.ID] {
			writeFolder(childFolder, depth+1)
		}

		sb.WriteString(fmt.Sprintf("%s</DL><p>\n", indent))
	}

	// 写入根文件夹
	for _, folder := range rootFolders {
		writeFolder(folder, 1)
	}

	// 写入根目录下的书签
	for _, bm := range rootBookmarks {
		sb.WriteString(fmt.Sprintf("    <DT><A HREF=\"%s\" ADD_DATE=\"%s\">%s</A>\n",
			html.EscapeString(bm.URL),
			strconv.FormatInt(bm.CreatedAt.Unix(), 10),
			html.EscapeString(bm.Title)))
	}

	sb.WriteString("</DL><p>\n")

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"bookmarks_%s.html\"", collection.Name))
	c.Header("Content-Type", "text/html; charset=UTF-8")
	c.String(http.StatusOK, sb.String())
}

// ImportBookmarksHTML 从 Netscape/Chrome HTML 格式导入书签（解析为JSON返回给前端）
func (h *BookmarkHandler) ImportBookmarksHTML(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请上传书签文件"))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("读取文件失败"))
		return
	}

	type ParsedItem struct {
		Type     string        `json:"type"`
		Title    string        `json:"title"`
		URL      string        `json:"url,omitempty"`
		Children []interface{} `json:"children,omitempty"`
	}

	var rootItems []interface{}
	folderStack := []*ParsedItem{}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "<DT><H3") {
			title := extractTagContent(trimmed, "H3")
			folder := &ParsedItem{Type: "folder", Title: title, Children: []interface{}{}}

			if len(folderStack) > 0 {
				parent := folderStack[len(folderStack)-1]
				parent.Children = append(parent.Children, folder)
			} else {
				rootItems = append(rootItems, folder)
			}
			folderStack = append(folderStack, folder)
		} else if strings.HasPrefix(trimmed, "<DT><A ") {
			itemURL := extractAttr(trimmed, "HREF")
			title := extractTagContent(trimmed, "A")
			link := ParsedItem{Type: "link", Title: title, URL: itemURL}

			if len(folderStack) > 0 {
				parent := folderStack[len(folderStack)-1]
				parent.Children = append(parent.Children, link)
			} else {
				rootItems = append(rootItems, link)
			}
		} else if strings.HasPrefix(trimmed, "</DL>") {
			if len(folderStack) > 0 {
				folderStack = folderStack[:len(folderStack)-1]
			}
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(rootItems, "解析成功"))
}

func extractTagContent(s, tag string) string {
	startTag := "<" + tag
	startIdx := strings.Index(s, startTag)
	if startIdx == -1 {
		return ""
	}
	afterTag := s[startIdx+len(startTag):]
	gtIdx := strings.Index(afterTag, ">")
	if gtIdx == -1 {
		return ""
	}
	content := afterTag[gtIdx+1:]
	endTag := "</" + tag + ">"
	endIdx := strings.Index(content, endTag)
	if endIdx == -1 {
		return html.UnescapeString(content)
	}
	return html.UnescapeString(content[:endIdx])
}

func extractAttr(s, attr string) string {
	search := attr + "=\""
	idx := strings.Index(s, search)
	if idx == -1 {
		return ""
	}
	valStart := idx + len(search)
	valEnd := strings.Index(s[valStart:], "\"")
	if valEnd == -1 {
		return ""
	}
	return s[valStart : valStart+valEnd]
}

// BatchDeleteBookmarks 批量删除书签
func (h *BookmarkHandler) BatchDeleteBookmarks(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请选择要删除的书签"))
		return
	}

	var bookmarks []models.Bookmark
	if err := internal.DB.Where("id IN ?", req.IDs).Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("查询书签失败"))
		return
	}

	// 删除关联的快照文件
	for _, bm := range bookmarks {
		if bm.SnapshotURL != "" {
			if err := h.snapshotService.DeleteSnapshot(bm.SnapshotURL); err != nil {
				log.Printf("删除快照文件失败 [书签ID: %s]: %v", bm.ID, err)
			}
		}
	}

	if err := internal.DB.Where("id IN ?", req.IDs).Delete(&models.Bookmark{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("批量删除失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, fmt.Sprintf("成功删除 %d 个书签", len(req.IDs))))
}

// BatchMoveBookmarks 批量移动书签到指定文件夹
func (h *BookmarkHandler) BatchMoveBookmarks(c *gin.Context) {
	var req struct {
		IDs      []string `json:"ids" binding:"required"`
		FolderID *string  `json:"folderId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请选择要移动的书签"))
		return
	}

	if err := internal.DB.Model(&models.Bookmark{}).Where("id IN ?", req.IDs).Select("folder_id").Updates(map[string]interface{}{
		"folder_id": req.FolderID,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("批量移动失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, fmt.Sprintf("成功移动 %d 个书签", len(req.IDs))))
}

// BatchUpdateSortOrders 批量更新书签排序
func (h *BookmarkHandler) BatchUpdateSortOrders(c *gin.Context) {
	var req struct {
		Orders []struct {
			ID    string `json:"id"`
			Order int    `json:"order"`
		} `json:"orders" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("参数错误"))
		return
	}

	tx := internal.DB.Begin()
	for _, item := range req.Orders {
		if err := tx.Model(&models.Bookmark{}).Where("id = ?", item.ID).Update("sort_order", item.Order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新排序失败"))
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "排序更新成功"))
}
