package handler

import "github.com/gin-gonic/gin"

// ok 统一成功响应
func ok(c *gin.Context, data any) {
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": data})
}

// fail 统一失败响应
func fail(c *gin.Context, httpStatus int, msg string) {
	c.JSON(httpStatus, gin.H{"code": httpStatus, "message": msg, "data": nil})
}
