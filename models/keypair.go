package models

import "gorm.io/gorm"

type KeyPair struct {
	gorm.Model
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Variant    string `json:"variant"`
}

// TableName overrides the table name used by KeyPair to `key_pairs`.
func (KeyPair) TableName() string {
	return "key_pairs"
}
