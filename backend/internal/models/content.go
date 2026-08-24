package models

import (
	"time"

	"github.com/google/uuid"
)

// BlogPost represents an article in the public blog.
// Shared across the platform (no tenant scope).
type BlogPost struct {
	Base
	Title           string     `gorm:"size:200;not null" json:"title"`
	Slug            string     `gorm:"size:200;uniqueIndex;not null" json:"slug"`
	Content         string     `gorm:"type:text;not null" json:"content"`
	Excerpt         *string    `gorm:"type:text" json:"excerpt,omitempty"`
	Category        string     `gorm:"size:100" json:"category"`
	CategoryColor   string     `gorm:"size:50" json:"category_color"`
	FeaturedImage   *string    `gorm:"size:512" json:"featured_image,omitempty"`
	Status          string     `gorm:"size:20;default:'draft'" json:"status"` // draft, published
	AuthorID        *uuid.UUID `gorm:"type:uuid" json:"author_id,omitempty"`  // Global author
	Author          *User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	MetaTitle       *string    `gorm:"size:150" json:"meta_title,omitempty"`
	MetaDescription *string    `gorm:"type:text" json:"meta_description,omitempty"`
}

// LegalDocument represents platform terms and policies.
type LegalDocument struct {
	Base
	Type          string     `gorm:"size:50;uniqueIndex;not null" json:"type"` // terms, privacy, cookie
	Title         string     `gorm:"size:200;not null" json:"title"`
	Content       string     `gorm:"type:text;not null" json:"content"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
	Version       string     `gorm:"size:20;default:'1.0'" json:"version"`
}

// FAQ represents frequently asked questions.
type FAQ struct {
	Base
	Question    string `gorm:"size:255" json:"question"`
	Answer      string `gorm:"type:text" json:"answer"`
	OrderIndex  uint   `gorm:"default:0" json:"order_index"`
	IsPublished bool   `gorm:"default:true" json:"is_published"`
}

// BlogCategory categorizes blog posts.
type BlogCategory struct {
	Base
	Name        string `gorm:"size:100" json:"name"`
	Slug        string `gorm:"size:100;uniqueIndex" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
}

// BlogTag tags blog posts.
type BlogTag struct {
	Base
	Name string `gorm:"size:100" json:"name"`
	Slug string `gorm:"size:100;uniqueIndex" json:"slug"`
}

// FeatureCategory groups product features.
type FeatureCategory struct {
	Base
	Name     string `gorm:"size:100" json:"name"`
	Icon     string `gorm:"size:50" json:"icon"`
	Order    uint   `gorm:"default:0" json:"order"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}

// FeatureItem represents a single feature in a category.
type FeatureItem struct {
	Base
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	Category   *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Title      string    `gorm:"size:200" json:"title"`
	Desc       string    `gorm:"type:text" json:"desc"`
	Icon       string    `gorm:"size:50" json:"icon"`
	Order      uint      `gorm:"default:0" json:"order"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	DetailsURL *string   `gorm:"size:255" json:"details_url,omitempty"`
}

// LeadershipMember represents team members on the about page.
type LeadershipMember struct {
	Base
	Name        string  `gorm:"size:150" json:"name"`
	Role        string  `gorm:"size:150" json:"role"`
	Image       *string `gorm:"size:512" json:"image,omitempty"`
	Bio         *string `gorm:"type:text" json:"bio,omitempty"`
	LinkedInURL *string `gorm:"size:512" json:"linkedin_url,omitempty"`
	TwitterURL  *string `gorm:"size:512" json:"twitter_url,omitempty"`
	Order       uint    `gorm:"default:0" json:"order"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
}

// DocumentationSection groups documentation articles.
type DocumentationSection struct {
	Base
	Title       string `gorm:"size:200" json:"title"`
	Slug        string `gorm:"size:200;uniqueIndex" json:"slug"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	Order       uint   `gorm:"default:0" json:"order"`
}

// DocumentationArticle is a single help page.
type DocumentationArticle struct {
	Base
	SectionID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"section_id"`
	Title       string     `gorm:"size:200" json:"title"`
	Slug        string     `gorm:"size:200;uniqueIndex" json:"slug"`
	Content     string     `gorm:"type:text" json:"content"`
	Order       uint       `gorm:"default:0" json:"order"`
	IsPublished bool       `gorm:"default:true" json:"is_published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}
