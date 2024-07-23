package repositories

import (
	"be-dilithium/models"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KeyPairRepository struct {
	DB *gorm.DB
}

func NewKeyPairRepository(db *gorm.DB) *KeyPairRepository {
	return &KeyPairRepository{DB: db}
}

func (r *KeyPairRepository) Create(ctx context.Context, keyPair *models.KeyPair) (*models.KeyPair, error) {
	if err := r.DB.WithContext(ctx).Create(keyPair).Error; err != nil {
		return nil, err
	}
	return keyPair, nil
}

func (r *KeyPairRepository) GetById(ctx context.Context, id uint) (*models.KeyPair, error) {
	var keyPair models.KeyPair
	if err := r.DB.WithContext(ctx).First(&keyPair, id).Error; err != nil {
		return nil, err
	}
	return &keyPair, nil
}

func (r *KeyPairRepository) GetAll(ctx context.Context) ([]models.KeyPair, error) {
	var keyPairs []models.KeyPair
	if err := r.DB.WithContext(ctx).Find(&keyPairs).Error; err != nil {
		return nil, err
	}
	return keyPairs, nil
}

func (r *KeyPairRepository) Update(ctx context.Context, keyPair *models.KeyPair) (*models.KeyPair, error) {
	if err := r.DB.WithContext(ctx).Save(keyPair).Error; err != nil {
		return nil, err
	}
	return keyPair, nil
}

func (r *KeyPairRepository) Delete(ctx context.Context, id uint) error {
	if err := r.DB.WithContext(ctx).Delete(&models.KeyPair{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *KeyPairRepository) FindByField(ctx context.Context, field string, value interface{}) ([]models.KeyPair, error) {
	var keyPairs []models.KeyPair
	query := r.DB.WithContext(ctx)

	if field != "" && value != nil {
		query = query.Where(clause.Eq{Column: field, Value: value})
	}

	if err := query.Find(&keyPairs).Error; err != nil {
		return nil, err
	}
	return keyPairs, nil
}
