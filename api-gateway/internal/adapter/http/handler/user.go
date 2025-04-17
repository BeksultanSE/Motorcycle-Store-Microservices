package handler

import (
	proto "github.com/BeksultanSE/Assignment1-api-gateway/pkg/protos/gen/golang"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"net/http"
)

func (h *Handler) RegisterUser(c *gin.Context) {
	var req proto.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := h.Clients.User.RegisterUser(c.Request.Context(), &req)
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

func (h *Handler) GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	req := &proto.UserID{UserId: userID.(uint64)}
	resp, err := h.Clients.User.GetUserProfile(c.Request.Context(), req)
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
