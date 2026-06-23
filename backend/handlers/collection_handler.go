package handlers

import (
	"log"
	"net/http"
	"pintree-backend/internal"
	"pintree-backend/models"
	"pintree-backend/utils"
	"github.com/gin-gonic/gin"
)

// CollectionHandler 收藏夹处理器
type CollectionHandler struct{}

func NewCollectionHandler() *CollectionHandler {
	return &CollectionHandler{}
}

// GetCollections 获取收藏夹列表
func (h *CollectionHandler) GetCollections(c *gin.Context) {
	publicOnly := c.Query("publicOnly") == "true"

	var collections []models.Collection
	query := internal.DB.Preload("User")

	// 获取用户ID（可能为空，因为使用了可选认证中间件）
	userID := c.GetString("userID")

	if publicOnly {
		// 明确要求只获取公开收藏夹
		query = query.Where("is_public = ?", true)
	} else if userID != "" {
		// 已登录用户，返回用户的收藏夹和公共收藏夹
		query = query.Where("is_public = ? OR user_id = ?", true, userID)
	} else {
		// 未登录用户，只返回公开收藏夹
		query = query.Where("is_public = ?", true)
	}

	if err := query.Order("sort_order ASC, created_at DESC").Find(&collections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取收藏夹失败"))
		return
	}

	// 填充每个收藏夹的书签数量
	for i := range collections {
		var count int64
		internal.DB.Model(&models.Bookmark{}).Where("collection_id = ?", collections[i].ID).Count(&count)
		collections[i].BookmarkCount = int(count)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(collections, ""))
}

// GetCollectionByID 根据ID获取收藏夹
func (h *CollectionHandler) GetCollectionByID(c *gin.Context) {
	id := c.Param("id")

	var collection models.Collection
	if err := internal.DB.Preload("User").First(&collection, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("收藏夹不存在"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(collection, ""))
}

// GetCollectionBySlug 根据slug获取收藏夹
func (h *CollectionHandler) GetCollectionBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var collection models.Collection
	if err := internal.DB.Preload("User").Where("slug = ?", slug).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("收藏夹不存在"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(collection, ""))
}

// CreateCollection 创建收藏夹
func (h *CollectionHandler) CreateCollection(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		IsPublic    bool   `json:"isPublic"`
		SortOrder   int    `json:"sortOrder"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}

	// 生成slug
	slug := utils.GenerateSlug(req.Name)

	collection := models.Collection{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Icon:        req.Icon,
		IsPublic:    req.IsPublic,
		SortOrder:   req.SortOrder,
		UserID:      userID,
	}

	if err := internal.DB.Create(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("创建收藏夹失败"))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse(collection, "创建成功"))
}

// UpdateCollection 更新收藏夹
func (h *CollectionHandler) UpdateCollection(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var collection models.Collection
	if err := internal.DB.First(&collection, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("收藏夹不存在"))
		return
	}

	// 检查权限
	if collection.UserID != userID {
		c.JSON(http.StatusForbidden, utils.ErrorResponse("无权操作"))
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		IsPublic    bool   `json:"isPublic"`
		SortOrder   int    `json:"sortOrder"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"icon":        req.Icon,
		"is_public":   req.IsPublic,
		"sort_order":  req.SortOrder,
	}

	if req.Name != "" && req.Name != collection.Name {
		updates["slug"] = utils.GenerateSlug(req.Name)
	}

	if err := internal.DB.Model(&collection).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(collection, "更新成功"))
}

// DeleteCollection 删除收藏夹
func (h *CollectionHandler) DeleteCollection(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var collection models.Collection
	if err := internal.DB.First(&collection, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("收藏夹不存在"))
		return
	}

	// 检查权限
	if collection.UserID != userID {
		c.JSON(http.StatusForbidden, utils.ErrorResponse("无权操作"))
		return
	}

	// 删除相关的文件夹和书签
	if err := internal.DB.Where("collection_id = ?", id).Delete(&models.Folder{}).Error; err != nil {
		log.Printf("删除收藏夹关联文件夹失败 [collectionID: %s]: %v", id, err)
	}
	if err := internal.DB.Where("collection_id = ?", id).Delete(&models.Bookmark{}).Error; err != nil {
		log.Printf("删除收藏夹关联书签失败 [collectionID: %s]: %v", id, err)
	}

	if err := internal.DB.Delete(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("删除失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "删除成功"))
}

// BatchUpdateSortOrders 批量更新收藏夹排序
func (h *CollectionHandler) BatchUpdateSortOrders(c *gin.Context) {
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
		if err := tx.Model(&models.Collection{}).Where("id = ?", item.ID).Update("sort_order", item.Order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新排序失败"))
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "排序更新成功"))
}
