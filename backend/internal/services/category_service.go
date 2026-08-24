package services

import (
	"errors"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

// CategoryWithCount embeds Category and adds a product count.
type CategoryWithCount struct {
	models.Category
	ProductCount int64 `json:"product_count"`
}

func (s *CategoryService) ListCategories() ([]CategoryWithCount, error) {
	var categories []models.Category
	if err := s.db.Find(&categories).Error; err != nil {
		return nil, err
	}

	result := make([]CategoryWithCount, 0, len(categories))
	for _, cat := range categories {
		var count int64
		s.db.Model(&models.Product{}).Where("category_id = ?", cat.ID).Count(&count)
		result = append(result, CategoryWithCount{Category: cat, ProductCount: count})
	}
	return result, nil
}

type CategoryCreateInput struct {
	Name        string
	Description string
	Color       string
}

func (s *CategoryService) CreateCategory(input CategoryCreateInput) (*models.Category, error) {
	category := models.Category{
		Name:        input.Name,
		Description: input.Description,
	}
	if input.Color != "" {
		category.Color = input.Color
	}

	if err := s.db.Create(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (s *CategoryService) GetCategory(id string) (*models.Category, error) {
	var category models.Category
	if err := s.db.Where("id = ?", id).First(&category).Error; err != nil {
		return nil, errors.New("category not found")
	}
	return &category, nil
}

type CategoryUpdateInput struct {
	Name        string
	Description string
	Color       string
}

func (s *CategoryService) UpdateCategory(id string, input CategoryUpdateInput) (*models.Category, error) {
	category, err := s.GetCategory(id)
	if err != nil {
		return nil, err
	}

	category.Name = input.Name
	category.Description = input.Description
	if input.Color != "" {
		category.Color = input.Color
	}

	if err := s.db.Save(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) DeleteCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.Category{}).Error
}
