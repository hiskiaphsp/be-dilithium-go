// repositories/signature_repository.go

package repositories

import (
	"be-dilithium/models"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SignatureRepository struct {
	DB *gorm.DB
}

func NewSignatureRepository(db *gorm.DB) *SignatureRepository {
	return &SignatureRepository{
		DB: db,
	}
}

func (r *SignatureRepository) Create(ctx context.Context, signature *models.Signature) (*models.Signature, error) {
	if err := r.DB.WithContext(ctx).Create(signature).Error; err != nil {
		return nil, err
	}
	return signature, nil
}

func (r *SignatureRepository) GetById(ctx context.Context, id uint) (*models.Signature, error) {
	var signature models.Signature
	if err := r.DB.WithContext(ctx).First(&signature, id).Error; err != nil {
		return nil, err
	}
	return &signature, nil
}

func (r *SignatureRepository) GetAll(ctx context.Context) ([]models.Signature, error) {
	var signatures []models.Signature
	if err := r.DB.WithContext(ctx).Find(&signatures).Error; err != nil {
		return nil, err
	}
	return signatures, nil
}

func (r *SignatureRepository) Update(ctx context.Context, signature *models.Signature) (*models.Signature, error) {
	if err := r.DB.WithContext(ctx).Save(signature).Error; err != nil {
		return nil, err
	}
	return signature, nil
}

func (r *SignatureRepository) Delete(ctx context.Context, id uint) error {
	if err := r.DB.WithContext(ctx).Delete(&models.Signature{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *SignatureRepository) FindByField(ctx context.Context, field string, value interface{}) ([]models.Signature, error) {
	var signatures []models.Signature
	query := r.DB.WithContext(ctx)

	if field != "" && value != nil {
		query = query.Where(clause.Eq{Column: field, Value: value})
	}

	if err := query.Find(&signatures).Error; err != nil {
		return nil, err
	}
	return signatures, nil
}

func (r *SignatureRepository) FindByMessageHashAndKeyId(ctx context.Context, documentHash string, keyID uint) ([]models.Signature, error) {
	var signatures []models.Signature
	if err := r.DB.WithContext(ctx).
		Where("document_hash = ? AND key_id = ?", documentHash, keyID).
		Find(&signatures).Error; err != nil {
		return nil, err
	}
	return signatures, nil
}
