package handler

import (
	"errors"
	"net/http"
	"toll_plaza/internal/domain"
	"toll_plaza/internal/service"

	"github.com/gin-gonic/gin"
)

type TollHandler struct {
	tollService service.TollService
}

func NewTollHandler(tollService service.TollService) *TollHandler {
	return &TollHandler{tollService: tollService}
}

// GetTollPlazas handles POST /api/v1/toll-plazas
func (h *TollHandler) GetTollPlazas(c *gin.Context) {
	var req domain.TollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "Invalid source or destination pincode",
		})
		return
	}

	resp, err := h.tollService.GetTollsBetweenPincodes(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidPincode):
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error: "Invalid source or destination pincode",
			})
		case errors.Is(err, domain.ErrSamePincode):
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error: "Source and destination pincodes cannot be the same",
			})
		case errors.Is(err, domain.ErrRouteNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Error: "Route between pincodes could not be calculated",
			})
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
