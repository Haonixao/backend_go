package public

import (
	"time"

	"backend_go/pkg/gorm_extra"
)

type Backup struct {
	gorm_extra.UuidId
	CreatedAt        time.Time `gorm:"type:timestamp with time zone;not null" json:"created_at"`
	Size             string    `gorm:"type:text;not null" json:"size"`
	MigrationVersion string    `gorm:"type:text;not null" json:"migration_version"`
	Data             []byte    `gorm:"type:bytea;not null" json:"data"`
}

func (b *Backup) TableName() string {
	return "public.backup"
}
