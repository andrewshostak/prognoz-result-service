package handler

import (
	"errors"
	"net/http"

	"github.com/andrewshostak/result-service/errs"
	"github.com/gin-gonic/gin"
)

type MatchHandler struct {
	matchService MatchService
}

func NewMatchHandler(matchService MatchService) *MatchHandler {
	return &MatchHandler{matchService: matchService}
}

func (h *MatchHandler) Create(c *gin.Context) {
	var params CreateMatchRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	result, err := h.matchService.Create(c.Request.Context(), params.ToDomain())
	if errors.As(err, &errs.AliasNotFoundError{}) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if errors.As(err, &errs.UnexpectedNumberOfItemsError{}) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"match_id": result})
}
