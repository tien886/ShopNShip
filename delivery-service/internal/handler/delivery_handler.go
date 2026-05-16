package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tien886/ShopNShip/delivery-service/internal/dto"
	"github.com/tien886/ShopNShip/delivery-service/internal/service"
)

type DeliveryHandler struct {
	svc service.DeliveryService
}

func NewDeliveryHandler(svc service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{svc: svc}
}

// GetDeliveries godoc
// @Summary      List deliveries
// @Description  Get all deliveries for the authenticated user (admin sees all)
// @Tags         Deliveries
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Delivery
// @Failure      500  {object}  map[string]string
// @Router       /deliveries [get]
func (h *DeliveryHandler) GetDeliveries(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		deliveries, err := h.svc.GetAllDeliveries()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, deliveries)
		return
	}

	deliveries, err := h.svc.GetUserDeliveries(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deliveries)
}

// GetDelivery godoc
// @Summary      Get a delivery by ID
// @Description  Retrieve a specific delivery by UUID
// @Tags         Deliveries
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Delivery UUID"
// @Success      200  {object}  model.Delivery
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /deliveries/{id} [get]
func (h *DeliveryHandler) GetDelivery(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery id"})
		return
	}

	delivery, err := h.svc.GetDelivery(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, delivery)
}

// UpdateStatus godoc
// @Summary      Update delivery status
// @Description  Update the status of a delivery (PENDING | ASSIGNED | IN_TRANSIT | DELIVERED | CANCELLED)
// @Tags         Deliveries
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                 true  "Delivery UUID"
// @Param        request  body      dto.UpdateStatusRequest  true  "New Status"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /deliveries/{id}/status [patch]
func (h *DeliveryHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery id"})
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}
