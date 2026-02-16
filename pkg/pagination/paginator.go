package pagination

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type PaginationConfig struct {
	DefaultPage  int
	DefaultLimit int
}

func DefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		DefaultPage:  1,
		DefaultLimit: 10,
	}
}

type PaginationRequest struct {
	Page  int `json:"page" query:"page" form:"page"`
	Limit int `json:"limit" query:"limit" form:"limit"`
}

type PaginatedResponse[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

func NewPaginatedResponse[T any](items []T, total int64, page, limit int) *PaginatedResponse[T] {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &PaginatedResponse[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ParsePagination parses pagination parameters from the request context.
func ParsePagination(ctx *fiber.Ctx, config ...PaginationConfig) *PaginationRequest {
	cfg := DefaultPaginationConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	req := &PaginationRequest{
		Page:  cfg.DefaultPage,
		Limit: cfg.DefaultLimit,
	}

	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if req.Page == cfg.DefaultPage { // Only if not set by query
		if pageStr := ctx.FormValue("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
				req.Page = page
			}
		}
	}

	if req.Limit == cfg.DefaultLimit { // Only if not set by query
		if limitStr := ctx.FormValue("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				req.Limit = limit
			}
		}
	}

	// Handle raw request body (JSON, form-urlencoded, or plain text)
	if req.Page == cfg.DefaultPage || req.Limit == cfg.DefaultLimit {
		contentType := ctx.Get("Content-Type")

		// Handle JSON body
		if contentType == "application/json" {
			bodyReq := &PaginationRequest{}
			if err := ctx.BodyParser(bodyReq); err == nil {
				if req.Page == cfg.DefaultPage && bodyReq.Page > 0 {
					req.Page = bodyReq.Page
				}
				if req.Limit == cfg.DefaultLimit && bodyReq.Limit > 0 {
					req.Limit = bodyReq.Limit
				}
			}
		}

		// Handle form-urlencoded body
		if contentType == "application/x-www-form-urlencoded" {
			bodyReq := &PaginationRequest{}
			if err := ctx.BodyParser(bodyReq); err == nil {
				if req.Page == cfg.DefaultPage && bodyReq.Page > 0 {
					req.Page = bodyReq.Page
				}
				if req.Limit == cfg.DefaultLimit && bodyReq.Limit > 0 {
					req.Limit = bodyReq.Limit
				}
			}
		}

		// Handle raw text body (e.g., "page=2&limit=20")
		if contentType == "text/plain" || contentType == "" {
			body := string(ctx.Body())
			if body != "" {
				// Parse raw text as query string format
				if pageStr := extractParam(body, "page"); pageStr != "" {
					if page, err := strconv.Atoi(pageStr); err == nil && page > 0 && req.Page == cfg.DefaultPage {
						req.Page = page
					}
				}
				if limitStr := extractParam(body, "limit"); limitStr != "" {
					if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && req.Limit == cfg.DefaultLimit {
						req.Limit = limit
					}
				}
			}
		}
	}

	req.Validate(cfg)

	return req
}

func (pr *PaginationRequest) Validate(config PaginationConfig) {
	if pr.Page <= 0 {
		pr.Page = config.DefaultPage
	}
	if pr.Limit <= 0 {
		pr.Limit = config.DefaultLimit
	}
}

func (pr *PaginationRequest) GetOffset() int {
	return (pr.Page - 1) * pr.Limit
}

func (pr *PaginationRequest) GetLimit() int {
	return pr.Limit
}

// extractParam extracts a parameter value from a query string format
func extractParam(body, param string) string {
	parts := strings.Split(body, "&")
	for _, part := range parts {
		if kv := strings.Split(part, "="); len(kv) == 2 && kv[0] == param {
			return kv[1]
		}
	}
	return ""
}
