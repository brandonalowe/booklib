package api

import (
	"booklib/internal/models"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/api/books/v1"
	"google.golang.org/api/option"
)

// searchGoogleApi searches for book metadata by ISBN using Google Books API.
// It accepts a context so callers can cancel or set timeouts. Returns a
// populated IsbnCache pointer on success, (nil, nil) if not found, or an
// error if the lookup failed.
func searchGoogleApi(ctx context.Context, isbn int) (*models.IsbnCache, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	apiKey := os.Getenv("GOOGLE_BOOK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_BOOK_API_KEY not set")
	}

	svc, err := books.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("isbn:%d", isbn)
	resp, err := svc.Volumes.List(q).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Items) == 0 {
		return nil, nil
	}

	v := resp.Items[0]
	vi := v.VolumeInfo

	var authors string
	if len(vi.Authors) > 0 {
		authors = strings.Join(vi.Authors, ", ")
	}

	var genre string
	if len(vi.Categories) > 0 {
		genre = vi.Categories[0]
	}

	var cover string
	if vi.ImageLinks != nil {
		if vi.ImageLinks.Thumbnail != "" {
			cover = vi.ImageLinks.Thumbnail
		} else if vi.ImageLinks.SmallThumbnail != "" {
			cover = vi.ImageLinks.SmallThumbnail
		}
	}

	now := time.Now()
	return &models.IsbnCache{
		ISBN:     fmt.Sprintf("%d", isbn),
		Title:    vi.Title,
		Author:   authors,
		Genre:    genre,
		CoverUrl: cover,
		CachedAt: &now,
	}, nil
}
