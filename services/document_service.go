// services/document_service.go

package services

import (
	"context"

	"be-dilithium/models"
	"be-dilithium/repositories"
)

type DocumentService struct {
	Repo *repositories.DocumentRepository
}

func NewDocumentService(repo *repositories.DocumentRepository) *DocumentService {
	return &DocumentService{Repo: repo}
}

func (s *DocumentService) Create(ctx context.Context, document *models.Document) (*models.Document, error) {
	_, err := s.Repo.Create(ctx, document)
	if err != nil {
		return nil, err
	}
	return document, nil
}

func (s *DocumentService) GetById(ctx context.Context, id string) (*models.Document, error) {
	return s.Repo.GetById(ctx, id)
}

func (s *DocumentService) GetAll(ctx context.Context) ([]models.Document, error) {
	return s.Repo.GetAll(ctx)
}

func (s *DocumentService) Update(ctx context.Context, document *models.Document) (*models.Document, error) {
	_, err := s.Repo.Update(ctx, document)
	if err != nil {
		return nil, err
	}
	return document, nil
}

func (s *DocumentService) Delete(ctx context.Context, id string) (bool, error) {
	_, err := s.Repo.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
