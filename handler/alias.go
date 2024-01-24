package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AliasHandler struct {
	aliasService AliasService
}

func NewAliasHandler(aliasService AliasService) *AliasHandler {
	return &AliasHandler{aliasService: aliasService}
}

func (h *AliasHandler) Search(c *gin.Context) {
	var params SearchAliasRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	result, err := h.aliasService.Search(c.Request.Context(), params.Search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"aliases": result})
}
