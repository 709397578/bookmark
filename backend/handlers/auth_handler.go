package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"pintree-backend/config"
	"pintree-backend/internal"
	"pintree-backend/models"
	"pintree-backend/utils"
	"unicode"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证相关处理器
type AuthHandler struct {
	config *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{config: cfg}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"max=50"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}

	var user models.User
	if err := internal.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("邮箱或密码错误"))
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("邮箱或密码错误"))
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role, h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("生成令牌失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
			"avatar": user.Avatar,
		},
	}, "登录成功"))
}

// validatePassword 校验密码强度
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度至少8位")
	}
	var hasLetter, hasDigit bool
	for _, ch := range password {
		if unicode.IsLetter(ch) {
			hasLetter = true
		}
		if unicode.IsDigit(ch) {
			hasDigit = true
		}
	}
	if !hasLetter {
		return fmt.Errorf("密码必须包含至少一个字母")
	}
	if !hasDigit {
		return fmt.Errorf("密码必须包含至少一个数字")
	}
	return nil
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	// 检查是否允许注册
	var regSetting models.Setting
	if err := internal.DB.Where("key = ?", "allowRegistration").First(&regSetting).Error; err == nil {
		var allowed bool
		if err := json.Unmarshal([]byte(regSetting.Value), &allowed); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("注册设置异常"))
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, utils.ErrorResponse("注册功能已关闭"))
			return
		}
	}

	var req RegisterRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}

	// 校验密码强度
	if err := validatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	// 检查邮箱是否已存在
	var existingUser models.User
	if err := internal.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, utils.ErrorResponse("该邮箱已被注册"))
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("密码加密失败"))
		return
	}

	// 创建用户
	user := models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Role:     "user",
	}

	if err := internal.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("注册失败"))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
	}, "注册成功"))
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")
	
	var user models.User
	if err := internal.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("用户不存在"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
		"avatar": user.Avatar,
	}, ""))
}

// ListUsers 获取所有用户（管理员）
func (h *AuthHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := internal.DB.Order("created_at DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取用户列表失败"))
		return
	}

	var result []gin.H
	for _, u := range users {
		var bookmarkCount int64
		internal.DB.Model(&models.Bookmark{}).Where("collection_id IN (SELECT id FROM collections WHERE user_id = ?)", u.ID).Count(&bookmarkCount)
		var collectionCount int64
		internal.DB.Model(&models.Collection{}).Where("user_id = ?", u.ID).Count(&collectionCount)
		result = append(result, gin.H{
			"id":             u.ID,
			"email":          u.Email,
			"name":           u.Name,
			"role":           u.Role,
			"collectionCount": collectionCount,
			"bookmarkCount":  bookmarkCount,
			"createdAt":      u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(result, ""))
}

// UpdateUser 更新用户信息（管理员）
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := internal.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("用户不存在"))
		return
	}

	var req struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("参数错误"))
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Role != "" && (req.Role == "admin" || req.Role == "user") {
		updates["role"] = req.Role
	}

	if len(updates) > 0 {
		if err := internal.DB.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新失败"))
			return
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	}, "更新成功"))
}

// DeleteUser 删除用户（管理员）
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetString("userID")

	if id == currentUserID {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("不能删除自己的账号"))
		return
	}

	var user models.User
	if err := internal.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("用户不存在"))
		return
	}

	// 删除用户相关的书签和文件夹
	internal.DB.Where("collection_id IN (SELECT id FROM collections WHERE user_id = ?)", id).Delete(&models.Bookmark{})
	internal.DB.Where("collection_id IN (SELECT id FROM collections WHERE user_id = ?)", id).Delete(&models.Folder{})
	internal.DB.Where("user_id = ?", id).Delete(&models.Collection{})
	internal.DB.Delete(&user)

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "删除成功"))
}

// UploadAvatar 上传用户头像
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("userID")

	file, _, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请上传头像文件"))
		return
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("读取文件失败"))
		return
	}

	// 检查文件大小 (2MB)
	if len(content) > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("头像文件不能超过2MB"))
		return
	}

	// 保存到 uploads/avatars 目录
	avatarDir := "./uploads/avatars"
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("创建目录失败"))
		return
	}

	filename := fmt.Sprintf("avatar_%s.jpg", userID)
	filepath := filepath.Join(avatarDir, filename)
	if err := os.WriteFile(filepath, content, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("保存文件失败"))
		return
	}

	avatarURL := "/avatars/" + filename

	// 更新用户头像
	if err := internal.DB.Model(&models.User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新头像失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(gin.H{"avatar": avatarURL}, "头像上传成功"))
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// ChangePassword 修改当前用户密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("userID")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}

	// 校验新密码强度
	if err := validatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
		return
	}

	// 新旧密码不能相同
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("新密码不能与旧密码相同"))
		return
	}

	var user models.User
	if err := internal.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("用户不存在"))
		return
	}

	// 校验旧密码
	if !utils.CheckPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("旧密码错误"))
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("密码加密失败"))
		return
	}

	if err := internal.DB.Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("修改密码失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "密码修改成功"))
}

// DeleteAvatar 删除用户头像
func (h *AuthHandler) DeleteAvatar(c *gin.Context) {
	userID := c.GetString("userID")

	var user models.User
	if err := internal.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("用户不存在"))
		return
	}

	if user.Avatar != "" {
		// 删除文件
		oldPath := "." + user.Avatar
		if strings.HasPrefix(user.Avatar, "/avatars/") {
			os.Remove(oldPath)
		}
		// 清空数据库字段
		internal.DB.Model(&user).Update("avatar", "")
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "头像已删除"))
}
