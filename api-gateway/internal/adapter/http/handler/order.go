package handler

import (
	protos "github.com/BeksultanSE/Assignment1-api-gateway/pkg/protos/gen/golang"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var items []*protos.CreateOrderItem
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		log.Println(err)
		return
	}

	req := &protos.CreateOrderRequest{
		UserId: userID.(uint64),
		Items:  items,
	}

	resp, err := h.Clients.Order.CreateOrder(c.Request.Context(), req)
	if err != nil {
		code, msg := mapGRPCErrorToHTTP(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}

	jsonBytes, err := protojson.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonBytes)
}
func (h *Handler) GetOrder(c *gin.Context) {
	UserId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		log.Printf(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	req := &protos.GetOrderRequest{
		OrderId: orderID,
	}

	resp, err := h.Clients.Order.GetOrder(c.Request.Context(), req)
	if err != nil {
		code, msg := mapGRPCErrorToHTTP(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}

	if resp.UserId != UserId && UserId != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid order ID, not allowed for current user"})
		return
	}

	jsonBytes, err := protojson.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonBytes)
}

func (h *Handler) UpdateOrder(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req protos.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.OrderId = orderID

	resp, err := h.Clients.Order.UpdateOrder(c.Request.Context(), &req)
	if err != nil {
		code, msg := mapGRPCErrorToHTTP(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}

	jsonBytes, err := protojson.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonBytes)
}

func (h *Handler) GetOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit < 1 {
		limit = 10
	}

	req := &protos.ListOrdersRequest{
		UserId: userID.(uint64),
		Page:   page,
		Limit:  limit,
	}

	resp, err := h.Clients.Order.ListOrders(c.Request.Context(), req)
	if err != nil {
		code, msg := mapGRPCErrorToHTTP(err)
		c.JSON(code, gin.H{"error": msg})
		return
	}

	jsonBytes, err := protojson.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonBytes)
}
