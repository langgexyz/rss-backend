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
	svc                *service.FeedService
	minRefreshInterval int
}

func NewFeedHandler(svc *service.FeedService, minRefreshInterval int) *FeedHandler {
	return &FeedHandler{svc: svc, minRefreshInterval: minRefreshInterval}
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

func (h *FeedHandler) Refresh(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondErr(c, http.StatusBadRequest, "INVALID_ID", "无效的 ID")
		return
	}
	err = h.svc.RefreshFeed(uint(id), h.minRefreshInterval)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respondErr(c, http.StatusNotFound, "FEED_NOT_FOUND", "订阅源不存在")
		return
	}
	if errors.Is(err, service.ErrTooSoon) {
		respondErr(c, http.StatusTooManyRequests, "TOO_SOON", "刷新过于频繁，请稍后再试")
		return
	}
	if err != nil {
		respondErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "刷新任务已触发"})
}

func (h *FeedHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondErr(c, http.StatusBadRequest, "INVALID_ID", "无效的 ID")
		return
	}
	if err := h.svc.DeleteFeed(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondErr(c, http.StatusNotFound, "FEED_NOT_FOUND", "订阅源不存在")
			return
		}
		respondErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
