package services

import (
	"be-dilithium/models"
	"be-dilithium/repositories"
	"context"
)

type KeyPairService struct {
	Repo *repositories.KeyPairRepository
}

func NewKeyPairService(repo *repositories.KeyPairRepository) *KeyPairService {
	return &KeyPairService{Repo: repo}
}

func (s *KeyPairService) Create(ctx context.Context, keyPair *models.KeyPair) (*models.KeyPair, error) {
	return s.Repo.Create(ctx, keyPair)
}

func (s *KeyPairService) GetById(ctx context.Context, id uint) (*models.KeyPair, error) {
	return s.Repo.GetById(ctx, id)
}

func (s *KeyPairService) GetAll(ctx context.Context) ([]models.KeyPair, error) {
	return s.Repo.GetAll(ctx)
}

func (s *KeyPairService) Update(ctx context.Context, keyPair *models.KeyPair) (*models.KeyPair, error) {
	return s.Repo.Update(ctx, keyPair)
}

func (s *KeyPairService) Delete(ctx context.Context, id uint) (bool, error) {
	err := s.Repo.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *KeyPairService) FindByField(ctx context.Context, field string, value interface{}) ([]models.KeyPair, error) {
	return s.Repo.FindByField(ctx, field, value)
}
