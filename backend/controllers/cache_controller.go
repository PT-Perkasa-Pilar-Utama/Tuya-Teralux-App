package controllers

import (
	"net/http"
	"teralux_app/services"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

// CacheController handles cache-related operations
type CacheController struct {
	cache *services.BadgerService
}

// NewCacheController creates a new CacheController instance
func NewCacheController(cache *services.BadgerService) *CacheController {
	return &CacheController{cache: cache}
}

// FlushCache clears the entire cache
// @Summary Flush all cache
// @Description Remove all data from the cache storage
// @Tags Cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/cache/flush [delete]
func (ctrl *CacheController) FlushCache(c *gin.Context) {
	if ctrl.cache == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     "Cache service not initialized",
		})
		return
	}

	err := ctrl.cache.FlushAll()
	if err != nil {
		utils.LogError("Failed to flush cache: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     "Failed to flush cache",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "Cache flushed successfully",
	})
}
