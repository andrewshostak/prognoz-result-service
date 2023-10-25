package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MatchHandler struct{}

func NewMatchHandler() *MatchHandler {
	return &MatchHandler{}
}

func (h *MatchHandler) Create(c *gin.Context) {
	var params CreateMatchRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})

		return
	}

	c.Status(200)
}
