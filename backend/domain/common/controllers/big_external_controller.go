package controllers

import (
	"fmt"
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/services"

	"github.com/gin-gonic/gin"
)

type BigExternalController struct {
	bigSvc *services.BigExternalService
}

func NewBigExternalController(bigSvc *services.BigExternalService) *BigExternalController {
	return &BigExternalController{bigSvc: bigSvc}
}

// GetDeviceInfo fetches device and booking information from Big API
// @Summary Fetch device and booking info by MAC address
// @Description Fetches specific device information including booking_id, time_start, and time_stop directly from the Big API.
// @Tags 08. Common
// @Accept json
// @Produce json
// @Param mac_address path string true "MAC Address"
// @Success 200 {object} dtos.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 404 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/big/device/{mac_address} [get]
func (c *BigExternalController) GetDeviceInfo(ctx *gin.Context) {
	macAddress := ctx.Param("mac_address")
	if macAddress == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "MAC address is required",
		})
		return
	}

	info, err := c.bigSvc.GetDeviceInfoByMac(macAddress)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to fetch device info: " + err.Error(),
		})
		return
	}

	bookingID := fmt.Sprintf("%v", info["SDTGetRoomTeraluxBookingid"])
	if bookingID == "<nil>" {
		bookingID = ""
	}

	timeEnd := fmt.Sprintf("%v", info["SDTGetRoomTeraluxtimeendDate"])
	if timeEnd == "<nil>" {
		timeEnd = ""
	}

	timeStart := fmt.Sprintf("%v", info["SDTGetRoomTeraluxtimeStartDate"])
	if timeStart == "<nil>" {
		timeStart = ""
	}

	// According to the previous exploration there was a booking_time logic in summary_usecase.go
	// "SDTGetRoomTeraluxBookingtimeChar" containing "10:00 - 11:00"
	bookingTimeChar := fmt.Sprintf("%v", info["SDTGetRoomTeraluxBookingtimeChar"])

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Success",
		Data: map[string]interface{}{
			"booking_id":            bookingID,
			"time_start":            timeStart,
			"time_stop":             timeEnd,
			"raw_booking_time_char": bookingTimeChar,
		},
	})
}
