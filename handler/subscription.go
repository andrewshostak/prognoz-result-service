package handler

import "github.com/gin-gonic/gin"

type SubscriptionHandler struct{}

func NewSubscriptionHandler() *SubscriptionHandler {
	return &SubscriptionHandler{}
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	c.Status(200)
}
