package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rss-backend/internal/fetcher"
	"rss-backend/internal/handler"
	"rss-backend/internal/repository"
	"rss-backend/internal/service"
	"rss-backend/internal/testutil"
)

func buildTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.SetupMySQL(t)
	service.InitSemaphore(1)

	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)
	f := fetcher.New(5, 10)

	feedSvc := service.NewFeedService(feedRepo, artRepo, f)
	artSvc := service.NewArticleService(artRepo)
	ftSvc := service.NewFulltextService(artRepo)

	feedH := handler.NewFeedHandler(feedSvc, 300)
	artH := handler.NewArticleHandler(artSvc, ftSvc)

	r := gin.New()
	api := r.Group("/api")
	api.POST("/feeds", feedH.Create)
	api.GET("/feeds", feedH.List)
	api.GET("/feeds/:id", feedH.GetByID)
	api.POST("/feeds/:id/refresh", feedH.Refresh)
	api.DELETE("/feeds/:id", feedH.Delete)
	api.GET("/articles", artH.List)
	api.GET("/articles/:id", artH.GetByID)
	api.PATCH("/articles/:id", artH.Update)
	return r
}

func postJSON(r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestFeedHandler_Create_InvalidURL(t *testing.T) {
	r := buildTestRouter(t)
	w := postJSON(r, "/api/feeds", map[string]string{"url": "not-a-url"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "INVALID_REQUEST", resp["code"])
}

func TestFeedHandler_Create_ValidURL_Returns202(t *testing.T) {
	r := buildTestRouter(t)
	w := postJSON(r, "/api/feeds", map[string]string{"url": "https://example.com/rss.xml"})
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestFeedHandler_Create_DuplicateURL_Returns409(t *testing.T) {
	r := buildTestRouter(t)
	postJSON(r, "/api/feeds", map[string]string{"url": "https://example.com/rss.xml"})
	w := postJSON(r, "/api/feeds", map[string]string{"url": "https://example.com/rss.xml"})
	assert.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "FEED_ALREADY_EXISTS", resp["code"])
}

func TestFeedHandler_GetByID_NotFound(t *testing.T) {
	r := buildTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/feeds/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFeedHandler_Delete(t *testing.T) {
	r := buildTestRouter(t)
	// Create first
	w1 := postJSON(r, "/api/feeds", map[string]string{"url": "https://todelete.com/rss"})
	require.Equal(t, http.StatusAccepted, w1.Code)
	var feed map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &feed)
	id := int(feed["ID"].(float64))

	// Delete
	req := httptest.NewRequest(http.MethodDelete, "/api/feeds/"+strconv.Itoa(id), nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusNoContent, w2.Code)
}

func TestArticleHandler_Update_FieldWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.SetupMySQL(t)
	service.InitSemaphore(1)

	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)
	f := fetcher.New(5, 10)
	feedSvc := service.NewFeedService(feedRepo, artRepo, f)
	artSvc := service.NewArticleService(artRepo)
	ftSvc := service.NewFulltextService(artRepo)

	// Seed a feed and article directly via repositories
	feed, err := feedRepo.Create("https://test.com/rss")
	require.NoError(t, err)
	err = artRepo.Upsert(feed.ID, []fetcher.NormalizedArticle{
		{
			GUIDHash:    "1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd",
			Title:       "Test",
			Content:     "original content",
			PublishedAt: time.Now(),
		},
	})
	require.NoError(t, err)
	result, err := artRepo.List(repository.ArticleFilter{Page: 1, PageSize: 1})
	require.NoError(t, err)
	require.NotEmpty(t, result.Items)
	artID := result.Items[0].ID

	feedH := handler.NewFeedHandler(feedSvc, 300)
	artH := handler.NewArticleHandler(artSvc, ftSvc)
	r := gin.New()
	api := r.Group("/api")
	api.POST("/feeds", feedH.Create)
	api.PATCH("/articles/:id", artH.Update)

	// PATCH with is_starred (allowed) and content (not allowed)
	body, _ := json.Marshal(map[string]interface{}{"is_starred": true, "content": "hacked"})
	req := httptest.NewRequest(http.MethodPatch, "/api/articles/"+strconv.Itoa(int(artID)), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var art map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &art)
	assert.Equal(t, true, art["IsStarred"])
	assert.Equal(t, "original content", art["Content"])
}
