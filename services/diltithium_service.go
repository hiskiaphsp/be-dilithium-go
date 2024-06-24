package services

import (
	"errors"

	"github.com/cloudflare/circl/sign/dilithium"
)

type DilithiumService struct {
	mode dilithium.Mode
}

func NewDilithiumService(modeName string) (*DilithiumService, error) {
	mode := dilithium.ModeByName(modeName)
	if mode == nil {
		return nil, errors.New("invalid Dilithium mode")
	}
	return &DilithiumService{mode: mode}, nil
}

func (s *DilithiumService) GenerateKeyPair() ([]byte, []byte, error) {
	pk, sk, err := s.mode.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}
	return pk.Bytes(), sk.Bytes(), nil
}

func (s *DilithiumService) SignMessage(privateKeyBytes []byte, message []byte) ([]byte, error) {
	sk := s.mode.PrivateKeyFromBytes(privateKeyBytes)
	return s.mode.Sign(sk, message), nil
}

func (s *DilithiumService) VerifySignature(publicKeyBytes []byte, message []byte, signature []byte) (bool, error) {
	pk := s.mode.PublicKeyFromBytes(publicKeyBytes)
	return s.mode.Verify(pk, message, signature), nil
}
