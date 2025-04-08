package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	inventoryServiceURL = "http://localhost:8001/api/v1"
	orderServiceURL     = "http://localhost:8002/api/v1"
	validPassword       = "your-secure-password" // Replace with your actual password
)

func main() {
	r := gin.Default()

	SetupRoutes(r)

	fmt.Println("API Gateway running on :8000...")
	if err := r.Run(":8000"); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}

func SetupRoutes(r *gin.Engine) {
	// Inventory Service
	r.Any("/products/*any", proxy(inventoryServiceURL))
	r.Any("/products", proxy(inventoryServiceURL))

	// Order Service
	r.Any("/orders/*any", proxy(orderServiceURL))
	r.Any("/orders", proxy(orderServiceURL))

	// Catch-all for unmatched routes
	r.NoRoute(func(c *gin.Context) {
		log.Printf("No route matched: %s %s", c.Request.Method, c.Request.URL.String())
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
	})
}

func proxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader != validPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Build target URL
		url := target + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			url += "?" + c.Request.URL.RawQuery
		}

		// Log for debugging
		log.Printf("Forwarding %s to %s", c.Request.Method, url)

		// Create new request
		req, err := http.NewRequest(c.Request.Method, url, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy headers
		req.Header = c.Request.Header.Clone()

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error contacting %s: %v", url, err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach target service"})
			return
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Forward response
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}
