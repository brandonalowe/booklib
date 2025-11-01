package api

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"booklib/internal/models"

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

// SearchGoogleApi is the exported version that can be used by handlers.
func SearchGoogleApi(ctx context.Context, isbn int) (*models.IsbnCache, error) {
	return searchGoogleApi(ctx, isbn)
}

// SearchGoogleBooks performs a generic search (title, author, or ISBN) and returns multiple results.
func SearchGoogleBooks(ctx context.Context, query string) ([]*models.IsbnCache, error) {
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

	// Search with the user's query (Google Books API will intelligently search title, author, ISBN)
	resp, err := svc.Volumes.List(query).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Items) == 0 {
		return []*models.IsbnCache{}, nil
	}

	results := make([]*models.IsbnCache, 0, len(resp.Items))
	now := time.Now()

	for _, v := range resp.Items {
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

		// Extract ISBN (prefer ISBN_13, fallback to ISBN_10)
		var isbn string
		if len(vi.IndustryIdentifiers) > 0 {
			for _, id := range vi.IndustryIdentifiers {
				if id.Type == "ISBN_13" {
					isbn = id.Identifier
					break
				}
			}
			if isbn == "" {
				isbn = vi.IndustryIdentifiers[0].Identifier
			}
		}

		results = append(results, &models.IsbnCache{
			ISBN:     isbn,
			Title:    vi.Title,
			Author:   authors,
			Genre:    genre,
			CoverUrl: cover,
			CachedAt: &now,
		})
	}

	return results, nil
}
