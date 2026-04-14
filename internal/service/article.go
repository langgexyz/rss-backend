package service

import (
	"rss-backend/internal/repository"
)

type ArticleService struct {
	artRepo *repository.ArticleRepository
}

func NewArticleService(ar *repository.ArticleRepository) *ArticleService {
	return &ArticleService{artRepo: ar}
}

func (s *ArticleService) ListArticles(f repository.ArticleFilter) (*repository.ArticleListResult, error) {
	return s.artRepo.List(f)
}

// GetArticle returns article details and automatically marks it as read
func (s *ArticleService) GetArticle(id uint) (*repository.ArticleWithFeedTitle, error) {
	art, err := s.artRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if !art.IsRead {
		s.artRepo.Update(id, map[string]interface{}{"is_read": true})
		art.IsRead = true
	}
	return art, nil
}

// UpdateArticle updates article fields, restricting to is_read and is_starred
func (s *ArticleService) UpdateArticle(id uint, updates map[string]interface{}) (*repository.ArticleWithFeedTitle, error) {
	// Only allow updating is_read and is_starred fields
	allowed := map[string]interface{}{}
	if v, ok := updates["is_read"]; ok {
		allowed["is_read"] = v
	}
	if v, ok := updates["is_starred"]; ok {
		allowed["is_starred"] = v
	}

	// If no allowed fields, just return current article
	if len(allowed) == 0 {
		return s.artRepo.GetByID(id)
	}

	// Apply updates
	if _, err := s.artRepo.Update(id, allowed); err != nil {
		return nil, err
	}

	// Return updated article
	return s.artRepo.GetByID(id)
}
