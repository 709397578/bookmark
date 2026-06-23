package handlers

import (
	"log"
	"net/http"
	"encoding/json"
	"pintree-backend/internal"
	"pintree-backend/models"
	"pintree-backend/utils"
	"github.com/gin-gonic/gin"
)

// SettingHandler 设置处理器
type SettingHandler struct{}

func NewSettingHandler() *SettingHandler {
	return &SettingHandler{}
}

// GetSettings 获取所有设置
func (h *SettingHandler) GetSettings(c *gin.Context) {
	var settings []models.Setting
	
	if err := internal.DB.Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取设置失败"))
		return
	}

		// 转换为map格式
		settingsMap := make(map[string]interface{})
		for _, setting := range settings {
			switch setting.Type {
			case "number":
				var num float64
				if err := json.Unmarshal([]byte(setting.Value), &num); err != nil {
					settingsMap[setting.Key] = setting.Value
				} else {
					settingsMap[setting.Key] = num
				}
			case "boolean":
				var bool_val bool
				if err := json.Unmarshal([]byte(setting.Value), &bool_val); err != nil {
					bool_val = setting.Value == "true" || setting.Value == `"true"`
					settingsMap[setting.Key] = bool_val
				} else {
					settingsMap[setting.Key] = bool_val
				}
			case "json":
				var json_val interface{}
				if err := json.Unmarshal([]byte(setting.Value), &json_val); err != nil {
					settingsMap[setting.Key] = setting.Value
				} else {
					settingsMap[setting.Key] = json_val
				}
			default:
				if setting.Key == "allowRegistration" {
					bool_val := setting.Value == "true" || setting.Value == `"true"`
					settingsMap[setting.Key] = bool_val
				} else {
					settingsMap[setting.Key] = setting.Value
				}
			}
		}

		// 始终确保 allowRegistration 以 boolean 返回
		if _, exists := settingsMap["allowRegistration"]; !exists {
			settingsMap["allowRegistration"] = true
		}

		log.Printf("[GetSettings] allowRegistration=%v (type=%T)", settingsMap["allowRegistration"], settingsMap["allowRegistration"])

		c.JSON(http.StatusOK, utils.SuccessResponse(settingsMap, ""))
}

// GetSettingByKey 根据key获取设置
func (h *SettingHandler) GetSettingByKey(c *gin.Context) {
	key := c.Param("key")
	
	var setting models.Setting
	if err := internal.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("设置不存在"))
		return
	}

	var value interface{}
	switch setting.Type {
	case "number":
		var num float64
		if err := json.Unmarshal([]byte(setting.Value), &num); err != nil {
			value = setting.Value
		} else {
			value = num
		}
	case "boolean":
		var bool_val bool
		if err := json.Unmarshal([]byte(setting.Value), &bool_val); err != nil {
			bool_val = setting.Value == "true" || setting.Value == `"true"`
			value = bool_val
		} else {
			value = bool_val
		}
	case "json":
		var json_val interface{}
		if err := json.Unmarshal([]byte(setting.Value), &json_val); err != nil {
			value = setting.Value
		} else {
			value = json_val
		}
	default:
		if setting.Key == "allowRegistration" {
			value = setting.Value == "true" || setting.Value == `"true"`
		} else {
			value = setting.Value
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(gin.H{
		"key":   setting.Key,
		"value": value,
		"type":  setting.Type,
	}, ""))
}

// UpdateSetting 更新设置
func (h *SettingHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	log.Printf("[UpdateSetting] key=%s", key)

	var req struct {
		Value interface{} `json:"value"`
		Type  string      `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[UpdateSetting] 参数错误: %v", err)
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("请求参数错误"))
		return
	}
	log.Printf("[UpdateSetting] req.Value=%v, type=%T", req.Value, req.Value)

	var setting models.Setting
	if err := internal.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		log.Printf("[UpdateSetting] 记录不存在，将创建新记录")
		settingType := req.Type
		if settingType == "" {
			switch req.Value.(type) {
			case bool:
				settingType = "boolean"
			case float64, int:
				settingType = "number"
			default:
				settingType = "string"
			}
		}
		setting = models.Setting{
			Key:  key,
			Type: settingType,
		}
	} else {
		log.Printf("[UpdateSetting] 找到现有记录: ID=%s, Value=%s, Type=%s", setting.ID, setting.Value, setting.Type)
	}

	var valueStr string
	if req.Value != nil {
		bytes, _ := json.Marshal(req.Value)
		valueStr = string(bytes)
	}

	setting.Value = valueStr
	if setting.Key == "allowRegistration" {
		setting.Type = "boolean"
	} else if req.Type != "" {
		setting.Type = req.Type
	}
	log.Printf("[UpdateSetting] 准备保存: Value=%s, Type=%s, ID=%s", setting.Value, setting.Type, setting.ID)

	if setting.ID == "" {
		if err := internal.DB.Create(&setting).Error; err != nil {
			log.Printf("[UpdateSetting] 创建失败: %v", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("创建设置失败"))
			return
		}
	} else {
		if err := internal.DB.Save(&setting).Error; err != nil {
			log.Printf("[UpdateSetting] 保存失败: %v", err)
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("更新设置失败"))
			return
		}
	}

	// 验证保存结果
	var verify models.Setting
	if err := internal.DB.Where("key = ?", key).First(&verify).Error; err == nil {
		log.Printf("[UpdateSetting] 保存后验证: key=%s, Value=%s, Type=%s", verify.Key, verify.Value, verify.Type)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(setting, "更新成功"))
}

// InitSettings 初始化默认设置
func (h *SettingHandler) InitSettings(c *gin.Context) {
	defaultSettings := []models.Setting{
		{Key: "siteName", Value: `"Pintree"`, Type: "string"},
		{Key: "siteDescription", Value: `"个人书签导航网站"`, Type: "string"},
		{Key: "siteLogo", Value: `""`, Type: "string"},
		{Key: "footerText", Value: `"© 2024 Pintree. All rights reserved."`, Type: "string"},
		{Key: "allowRegistration", Value: `true`, Type: "boolean"},
	}

	for _, setting := range defaultSettings {
		var existing models.Setting
		if err := internal.DB.Where("key = ?", setting.Key).First(&existing).Error; err != nil {
			internal.DB.Create(&setting)
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(nil, "初始化成功"))
}

// GetPublicSettings 获取公开设置（无需认证）
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	var settings []models.Setting
	// 只返回公开需要的设置
	if err := internal.DB.Where("key IN ?", []string{"allowRegistration"}).Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("获取设置失败"))
		return
	}

		settingsMap := make(map[string]interface{})
		for _, setting := range settings {
			switch setting.Type {
			case "boolean":
				var boolVal bool
				if err := json.Unmarshal([]byte(setting.Value), &boolVal); err == nil {
					settingsMap[setting.Key] = boolVal
				} else {
					boolVal = setting.Value == "true" || setting.Value == `"true"`
					settingsMap[setting.Key] = boolVal
				}
			default:
				if setting.Key == "allowRegistration" {
					boolVal := setting.Value == "true" || setting.Value == `"true"`
					settingsMap[setting.Key] = boolVal
				} else {
					settingsMap[setting.Key] = setting.Value
				}
			}
		}

		// 如果没有设置过 allowRegistration，默认为 true
		if _, exists := settingsMap["allowRegistration"]; !exists {
			settingsMap["allowRegistration"] = true
		}

	c.JSON(http.StatusOK, utils.SuccessResponse(settingsMap, ""))
}
