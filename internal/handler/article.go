package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"rss-backend/internal/repository"
	"rss-backend/internal/service"
)

type ArticleHandler struct {
	svc         *service.ArticleService
	fulltextSvc *service.FulltextService
}

func NewArticleHandler(svc *service.ArticleService, ftSvc *service.FulltextService) *ArticleHandler {
	return &ArticleHandler{svc: svc, fulltextSvc: ftSvc}
}

func (h *ArticleHandler) List(c *gin.Context) {
	f := repository.ArticleFilter{
		Page:     1,
		PageSize: 20,
	}

	if v := c.Query("feed_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			uid := uint(id)
			f.FeedID = &uid
		}
	}
	if v := c.Query("starred"); v == "1" || v == "true" {
		b := true
		f.Starred = &b
	}
	if v := c.Query("unread"); v == "1" || v == "true" {
		b := true
		f.Unread = &b
	}
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		f.Page = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "20")); err == nil && ps > 0 && ps <= 100 {
		f.PageSize = ps
	}

	result, err := h.svc.ListArticles(f)
	if err != nil {
		respondErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondOK(c, gin.H{
		"total":     result.Total,
		"page":      f.Page,
		"page_size": f.PageSize,
		"items":     result.Items,
	})
}

func (h *ArticleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondErr(c, http.StatusBadRequest, "INVALID_ID", "无效的 ID")
		return
	}
	art, err := h.svc.GetArticle(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respondErr(c, http.StatusNotFound, "ARTICLE_NOT_FOUND", "文章不存在")
		return
	}
	if err != nil {
		respondErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondOK(c, art)
}

func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondErr(c, http.StatusBadRequest, "INVALID_ID", "无效的 ID")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErr(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	art, err := h.svc.UpdateArticle(uint(id), req)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respondErr(c, http.StatusNotFound, "ARTICLE_NOT_FOUND", "文章不存在")
		return
	}
	if err != nil {
		respondErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	respondOK(c, art)
}

func (h *ArticleHandler) FetchFulltext(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondErr(c, http.StatusBadRequest, "INVALID_ID", "无效的 ID")
		return
	}
	art, err := h.fulltextSvc.FetchFulltext(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respondErr(c, http.StatusNotFound, "ARTICLE_NOT_FOUND", "文章不存在")
		return
	}
	if err != nil {
		respondErr(c, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}
	respondOK(c, art)
}
