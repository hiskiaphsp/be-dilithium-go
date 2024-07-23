package models

import "gorm.io/gorm"

type Signature struct {
	gorm.Model
	KeyID        uint    `json:"key_id"`
	DocumentHash string  `json:"document_hash"`
	Signature    string  `json:"signature"`
	Key          KeyPair `gorm:"foreignKey:KeyID" json:"key"`
}

func (Signature) TableName() string {
	return "signatures"
}
