package utils

import (
	"fmt"
	"time"
)

// Response 统一响应格式
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}, message string) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string) Response {
	return Response{
		Success: false,
		Error:   message,
	}
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Data        interface{} `json:"data"`
	Total       int64       `json:"total"`
	CurrentPage int         `json:"currentPage"`
	TotalPages  int         `json:"totalPages"`
	PageSize    int         `json:"pageSize"`
}

// GenerateSlug 生成slug
func GenerateSlug(name string) string {
	slug := name
	// 转换为小写
	for i := 'A'; i <= 'Z'; i++ {
		slug = replaceChar(slug, string(i), string(i+32))
	}
	// 替换空格为连字符
	slug = replaceChar(slug, " ", "-")
	// 添加时间戳避免重复
	slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	return slug
}

func replaceChar(s, old, new string) string {
	result := ""
	for _, char := range s {
		if string(char) == old {
			result += new
		} else {
			result += string(char)
		}
	}
	return result
}
