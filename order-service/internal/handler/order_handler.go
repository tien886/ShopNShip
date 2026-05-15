package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tien886/ShopNShip/order-service/internal/dto"
	"github.com/tien886/ShopNShip/order-service/internal/service"
)

type OrderHandler struct {
	orderSvc service.OrderService
}

func NewOrderHandler(orderSvc service.OrderService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc}
}

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Create a new order for the authenticated user
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CreateOrderRequest  true  "Order Items"
// @Success      201      {object}  dto.OrderResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderSvc.CreateOrder(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrder godoc
// @Summary      Get an order by ID
// @Description  Retrieve a specific order by its UUID (only owner can access)
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Order UUID"
// @Success      200  {object}  dto.OrderResponse
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.orderSvc.GetOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Basic security check: user can only see their own orders
	// (unless they are admin, but we'll keep it simple for now)
	userID := c.MustGet("user_id").(uint)
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetUserOrders godoc
// @Summary      List orders for current user
// @Description  Retrieve all orders belonging to the authenticated user
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   dto.OrderResponse
// @Failure      500  {object}  map[string]string
// @Router       /orders [get]
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	orders, err := h.orderSvc.GetUserOrders(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// UpdateStatus godoc
// @Summary      Update order status
// @Description  Update the status of an order (PENDING | PAID | CANCELLED | COMPLETED)
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                   true  "Order UUID"
// @Param        request  body      dto.UpdateStatusRequest  true  "New Status"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderSvc.UpdateOrderStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}
