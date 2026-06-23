package handlers

import (
	"log"
	"net/http"
	"pintree-backend/internal"
	"pintree-backend/models"
	"pintree-backend/utils"

	"github.com/gin-gonic/gin"
)

// FolderHandler 文件夹处理器
type FolderHandler struct{}

func NewFolderHandler() *FolderHandler {
	return &FolderHandler{}
}

// GetFolders 获取文件夹列表
func (h *FolderHandler) GetFolders(c *gin.Context) {
	collectionID := c.Query("collectionId")

	var folders []models.Folder
	query := internal.DB.Preload("Collection")

	// 获取用户ID（可能为空，因为使用了可选认证中间件）
	userID := c.GetString("userID")

	if collectionID != "" {
		// 检查收藏夹是否公开
		var collection models.Collection
		if err := internal.DB.First(&collection, "id = ?", collectionID).Error; err == nil {
			// 如果收藏夹不公开且用户未登录或不是所有者，返回空列表
			if !collection.IsPublic && (userID == "" || userID != collection.UserID) {
				c.JSON(http.StatusOK, utils.SuccessResponse([]models.Folder{}, ""))
				return
			}
		}

		query = query.Where("collection_id = ?", collectionID)
	}

	if err := query.Order("sort_order ASC, created_at DESC").Find(&folders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取文件夹失败"))
		return
	}

	for i := range folders {
		var count int64
		internal.DB.Model(&models.Bookmark{}).Where("folder_id = ?", folders[i].ID).Count(&count)
		folders[i].BookmarkCount = int(count)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(folders, ""))
}

// GetFolderByID 根据ID获取文件夹
func (h *FolderHandler) GetFolderByID(c *gin.Context) {
	id := c.Param("id")

	var folder models.Folder
	if err := internal.DB.Preload("Collection").First(&folder, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("文件夹不存在"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(folder, ""))
}

// CreateFolder 创建文件夹
func (h *FolderHandler) CreateFolder(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		CollectionID string `json:"collectionId" binding:"required"`
		ParentID    *string `json:"parentId"`
		SortOrder   int    `json:"sortOrder"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}

	// 检查收藏夹是否存在
	var collection models.Collection
	if err := internal.DB.First(&collection, "id = ?", req.CollectionID).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("收藏夹不存在"))
		return
	}

	// 检查权限
	if collection.UserID != userID {
		c.JSON(http.StatusForbidden, utils.ErrorResponse("无权操作"))
		return
	}

	folder := models.Folder{
		Name:        req.Name,
		Icon:        req.Icon,
		Color:       req.Color,
		CollectionID: req.CollectionID,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
	}

	if err := internal.DB.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("创建文件夹失败"))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse(folder, "创建成功"))
}

// UpdateFolder 更新文件夹
func (h *FolderHandler) UpdateFolder(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var folder models.Folder
	if err := internal.DB.Preload("Collection").First(&folder, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("文件夹不存在"))
		return
	}

	// 检查权限
	if folder.Collection.ID == "" {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("文件夹关联的收藏夹不存在"))
		return
	}

	
	if folder.Collection.UserID != userID {
		c.JSON(http.StatusForbidden, utils.ErrorResponse("无权操作"))
		return
	}

	var req struct {
		Name      *string `json:"name"`
		Icon      *string `json:"icon"`
		Color     *string `json:"color"`
		ParentID  *string `json:"parentId"`
		SortOrder *int    `json:"sortOrder"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	// 处理颜色字段，允许设置为空值
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	// 不处理空的颜色字段，保留原有值
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if err := internal.DB.Model(&folder).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(folder, "更新成功"))
}

// DeleteFolder 删除文件夹
func (h *FolderHandler) DeleteFolder(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var folder models.Folder
	if err := internal.DB.Preload("Collection").First(&folder, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("文件夹不存在"))
		return
	}

	// 检查权限
	if folder.Collection.UserID != userID {
		c.JSON(http.StatusForbidden, utils.ErrorResponse("无权操作"))
		return
	}

	// 删除文件夹及其子文件夹
	if err := internal.DB.Where("parent_id = ?", id).Delete(&models.Folder{}).Error; err != nil {
		log.Printf("删除子文件夹失败 [folderID: %s]: %v", id, err)
	}
	// 删除文件夹中的书签
	if err := internal.DB.Where("folder_id = ?", id).Delete(&models.Bookmark{}).Error; err != nil {
		log.Printf("删除文件夹中书签失败 [folderID: %s]: %v", id, err)
	}
	// 删除文件夹
	if err := internal.DB.Delete(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("删除失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "删除成功"))
}

// BatchUpdateSortOrders 批量更新文件夹排序
func (h *FolderHandler) BatchUpdateSortOrders(c *gin.Context) {
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
		if err := tx.Model(&models.Folder{}).Where("id = ?", item.ID).Update("sort_order", item.Order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新排序失败"))
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "排序更新成功"))
}
