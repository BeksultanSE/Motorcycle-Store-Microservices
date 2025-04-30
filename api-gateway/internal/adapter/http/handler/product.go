package handler

import (
	proto "github.com/BeksultanSE/Assignment1-api-gateway/pkg/protos/gen/golang"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) CreateProduct(c *gin.Context) {
	var req proto.CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		log.Println(err)
		return
	}

	resp, err := h.Clients.Inventory.CreateProduct(c.Request.Context(), &req)
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

func (h *Handler) GetProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	req := &proto.GetProductRequest{
		ProductId: productID,
	}

	resp, err := h.Clients.Inventory.GetProduct(c.Request.Context(), req)
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

func (h *Handler) UpdateProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req proto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.ProductId = productID

	resp, err := h.Clients.Inventory.UpdateProduct(c.Request.Context(), &req)
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

func (h *Handler) ListProducts(c *gin.Context) {
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

	req := &proto.ListProductsRequest{
		Page:  page,
		Limit: limit,
	}

	resp, err := h.Clients.Inventory.ListProducts(c.Request.Context(), req)
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

func (h *Handler) DeleteProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	req := &proto.DeleteProductRequest{
		ProductId: productID,
	}

	resp, err := h.Clients.Inventory.DeleteProduct(c.Request.Context(), req)
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
