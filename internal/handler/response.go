package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondErr(c *gin.Context, status int, code, msg string) {
	c.JSON(status, ErrResponse{Code: code, Message: msg})
}

func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}
