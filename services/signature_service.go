// services/signature_service.go

package services

import (
	"be-dilithium/models"
	"be-dilithium/repositories"
	"context"
)

type SignatureService struct {
	Repo *repositories.SignatureRepository
}

func NewSignatureService(repo *repositories.SignatureRepository) *SignatureService {
	return &SignatureService{Repo: repo}
}

func (s *SignatureService) Create(ctx context.Context, signature *models.Signature) (*models.Signature, error) {
	sig, err := s.Repo.Create(ctx, signature)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignatureService) GetById(ctx context.Context, id uint) (*models.Signature, error) {
	return s.Repo.GetById(ctx, id)
}

func (s *SignatureService) GetAll(ctx context.Context) ([]models.Signature, error) {
	return s.Repo.GetAll(ctx)
}

func (s *SignatureService) Update(ctx context.Context, signature *models.Signature) (*models.Signature, error) {
	sig, err := s.Repo.Update(ctx, signature)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignatureService) Delete(ctx context.Context, id uint) (bool, error) {
	err := s.Repo.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SignatureService) FindByField(ctx context.Context, field string, value interface{}) ([]models.Signature, error) {
	return s.Repo.FindByField(ctx, field, value)
}

func (s *SignatureService) FindByMessageHashAndKeyId(ctx context.Context, documentHash string, keyID uint) ([]models.Signature, error) {
	return s.Repo.FindByMessageHashAndKeyId(ctx, documentHash, keyID)
}
