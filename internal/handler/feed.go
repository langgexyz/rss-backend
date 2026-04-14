package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"rss-backend/internal/service"
)

type FeedHandler struct {
	svc *service.FeedService
}

func NewFeedHandler(svc *service.FeedService) *FeedHandler {
	return &FeedHandler{svc: svc}
}

func (h *FeedHandler) Create(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErr(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	feed, err := h.svc.CreateFeed(req.URL)
	if err != nil {
		// MySQL duplicate entry: error number 1062
		if strings.Contains(err.Error(), "1062") || strings.Contains(err.Error(), "Duplicate entry") {
			respondErr(c, http.StatusConflict, "FEED_ALREADY_EXISTS", "该订阅源已添加")
			return
		}
		respondErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusAccepted, feed)
}

func (h *FeedHandler) List(c *gin.Context) {
	feeds, err := h.svc.ListFeeds()
	if err != nil {
		respondErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondOK(c, feeds)
}

func (h *FeedHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondErr(c, http.StatusBadRequest, "INVALID_ID", "无效的 ID")
		return
	}
	feed, err := h.svc.GetFeed(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respondErr(c, http.StatusNotFound, "FEED_NOT_FOUND", "订阅源不存在")
		return
	}
	if err != nil {
		respondErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondOK(c, feed)
}
