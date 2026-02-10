package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// CacheController handles cache-related operations
type CacheController struct {
	cache  *infrastructure.BadgerService
	vector *infrastructure.VectorService
}

// NewCacheController creates a new CacheController instance
func NewCacheController(cache *infrastructure.BadgerService, vector *infrastructure.VectorService) *CacheController {
	return &CacheController{
		cache:  cache,
		vector: vector,
	}
}

// FlushCache clears the entire cache
// @Summary Flush all cache
// @Description Remove all data from the cache storage
// @Tags         06. Flush
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/cache/flush [delete]
func (ctrl *CacheController) FlushCache(c *gin.Context) {
	if ctrl.cache == nil {
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Cache service not initialized",
			Data:    nil,
		})
		return
	}

	// 1. Flush BadgerDB Cache
	err := ctrl.cache.FlushAll()
	if err != nil {
		utils.LogError("Failed to flush badger cache: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to flush badger cache",
			Data:    nil,
		})
		return
	}

	// 2. Flush Vector DB Cache (if initialized)
	if ctrl.vector != nil {
		err = ctrl.vector.FlushAll()
		if err != nil {
			utils.LogError("Failed to flush vector cache: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
				Status:  false,
				Message: "Failed to flush vector cache",
				Data:    nil,
			})
			return
		}
	}

	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "All caches flushed successfully",
		Data:    nil,
	})
}
