package models

import "time"

type BlogCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type BlogPost struct {
	ID            int             `json:"id"`
	Title         string          `json:"title"`
	Slug          string          `json:"slug"`
	Excerpt       string          `json:"excerpt"`
	Content       string          `json:"content"`
	Thumbnail     string          `json:"thumbnail"`
	AuthorName    string          `json:"author_name"`
	CategoryID    int             `json:"category_id"`
	CategoryName  string          `json:"category_name"`
	Status        string          `json:"status"`
	PublishedAt   string          `json:"published_at"`
	CreatedAt     time.Time       `json:"created_at"`
	RelatedWisata []PopularWisata `json:"related_wisata,omitempty"`
}
