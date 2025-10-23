package models

import "time"

type IsbnCache struct {
	ISBN     string     `json:"isbn"`
	Title    string     `json:"title"`
	Author   string     `json:"author"`
	Genre    string     `json:"genre"`
	CoverUrl string     `json:"cover_url"`
	CachedAt *time.Time `json:"cached_at"`
}
