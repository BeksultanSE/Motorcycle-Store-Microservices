package middleware

import (
	proto "github.com/BeksultanSE/Assignment1-api-gateway/pkg/protos/gen/golang"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AuthMiddleware(userClient proto.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.GetHeader("X-Email")
		password := c.GetHeader("X-Password")

		if email == "" || password == "" {
			c.JSON(400, gin.H{"error": "X-Email and X-Password headers are required"})
			c.Abort()
			return
		}

		authReq := &proto.AuthRequest{
			Email:    email,
			Password: password,
		}

		authResp, err := userClient.AuthenticateUser(c.Request.Context(), authReq)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.Unauthenticated {
				c.JSON(401, gin.H{"error": "authentication failed"})
			} else {
				c.JSON(500, gin.H{"error": "internal server error"})
			}
			c.Abort()
			return
		}

		if !authResp.Authenticated {
			c.JSON(401, gin.H{"error": "authentication failed"})
			c.Abort()
			return
		}

		c.Set("user_id", authResp.UserId)
		c.Next()
	}
}
