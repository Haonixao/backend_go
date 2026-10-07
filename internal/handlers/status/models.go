package status

import (
	"time"

	"github.com/google/uuid"
)

type ResponseStatus struct {
	Service  string `json:"service"`
	Database string `json:"database"`
}

type Version struct {
	ID        string    `json:"id"`
	AppliedAt time.Time `json:"applied_at"`
}

type BackupDTO struct {
	ID               uuid.UUID `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	Size             string    `json:"size"`
	MigrationVersion string    `json:"migration_version"`
}

type ApplyBackupRequest struct {
	Strategy string `json:"strategy" binding:"required,oneof=skip replace fail" error:"стратегия должна быть: skip, replace, or fail"`
}

// Page
// Только для swagger. Структура должна совпадать с gorm_extra.Page[]
type Page[T any] struct {
	Page         int64   `json:"page"`
	Size         int64   `json:"size"`
	MaxPage      int64   `json:"max_page"`
	TotalPages   int64   `json:"total_pages"`
	Total        int64   `json:"total"`
	Last         bool    `json:"last"`
	First        bool    `json:"first"`
	Visible      int64   `json:"visible"`
	Error        *bool   `json:"-"`
	ErrorMessage *string `json:"-"`
	RawError     error   `json:"-"`
	Items        []*T    `json:"items"`
}
