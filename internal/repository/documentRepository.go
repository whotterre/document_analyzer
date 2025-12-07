package repository

import (
	"whotterre/doculyzer/internal/models"

	"gorm.io/gorm"
)

type DocumentRepository struct {
	DB *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{DB: db}
}

func (r *DocumentRepository) Create(doc *models.Document) error {
	return r.DB.Create(doc).Error
}

func (r *DocumentRepository) GetByID(id string) (*models.Document, error) {
	var doc models.Document
	if err := r.DB.First(&doc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *DocumentRepository) Update(doc *models.Document) error {
	return r.DB.Save(doc).Error
}
