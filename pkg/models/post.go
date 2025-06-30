package models

import "time"

type Post struct {
	ID          int        `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Content     string     `json:"content"`
	Images      string     `json:"images"`
	AuthorID    int        `json:"author_id"`
	Status      string     `json:"status"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	ModifiedAt  time.Time  `json:"modified_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}
