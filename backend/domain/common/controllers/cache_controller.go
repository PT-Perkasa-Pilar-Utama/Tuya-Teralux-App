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
// @Tags         07. Common
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/cache/flush [delete]
func (ctrl *CacheController) FlushCache(c *gin.Context) {
	if ctrl.cache == nil {
		utils.LogError("CacheController.FlushCache: Cache service not initialized")
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	// 1. Flush BadgerDB Cache
	err := ctrl.cache.FlushAll()
	if err != nil {
		utils.LogError("CacheController.FlushCache (Badger): %v", err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	// 2. Flush Vector DB Cache (if initialized)
	if ctrl.vector != nil {
		err = ctrl.vector.FlushAll()
		if err != nil {
			utils.LogError("CacheController.FlushCache (Vector): %v", err)
			c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
				Status:  false,
				Message: "Internal Server Error",
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
