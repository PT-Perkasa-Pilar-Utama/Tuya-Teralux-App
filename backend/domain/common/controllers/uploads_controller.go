package controllers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeProtectedUploads serves files from ./uploads with auth check
func ServeProtectedUploads() gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		safeName := filepath.Base(filename)
		if safeName == "" || strings.Contains(safeName, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}
		filePath := filepath.Join("./uploads", safeName)
		c.File(filePath)
	}
}
