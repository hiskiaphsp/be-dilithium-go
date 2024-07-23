// repositories/document_repository.go

package repositories

import (
	"be-dilithium/models"
	"context"

	"gorm.io/gorm"
)

type DocumentRepository struct {
	DB *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{
		DB: db,
	}
}

func (r *DocumentRepository) Create(ctx context.Context, document *models.Document) (*models.Document, error) {
	if err := r.DB.WithContext(ctx).Create(document).Error; err != nil {
		return nil, err
	}
	return document, nil
}

func (r *DocumentRepository) GetById(ctx context.Context, id uint) (*models.Document, error) {
	var document models.Document
	if err := r.DB.WithContext(ctx).First(&document, id).Error; err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) GetAll(ctx context.Context) ([]models.Document, error) {
	var documents []models.Document
	if err := r.DB.WithContext(ctx).Find(&documents).Error; err != nil {
		return nil, err
	}
	return documents, nil
}

func (r *DocumentRepository) Update(ctx context.Context, document *models.Document) (*models.Document, error) {
	if err := r.DB.WithContext(ctx).Save(document).Error; err != nil {
		return nil, err
	}
	return document, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id uint) error {
	if err := r.DB.WithContext(ctx).Delete(&models.Document{}, id).Error; err != nil {
		return err
	}
	return nil
}
