package models

import "gorm.io/gorm"

type Document struct {
	gorm.Model
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

func (Document) TableName() string {
	return "documents"
}
