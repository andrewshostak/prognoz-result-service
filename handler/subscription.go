package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct{}

func NewSubscriptionHandler() *SubscriptionHandler {
	return &SubscriptionHandler{}
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var params CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	c.Status(200)
}
