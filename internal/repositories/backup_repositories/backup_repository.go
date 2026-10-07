package backup_repositories

import (
	"time"

	"backend_go/pkg/format_errors"

	"backend_go/internal/models/public"
	"backend_go/pkg/gorm_extra"

	"github.com/google/uuid"
)

type BackupRepository interface {
	gorm_extra.BaseRepository[public.Backup, uuid.UUID]
	DeleteOldBackups(beforeDate time.Time) error
}

type BackupRepositoryImpl struct {
	gorm_extra.BaseRepository[public.Backup, uuid.UUID]
}

func (r *BackupRepositoryImpl) DeleteOldBackups(beforeDate time.Time) error {
	conn, err := r.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	query := conn.Where("created_at < ?", beforeDate)
	err = r.DeleteWithTx(query)
	if err != nil {
		err = format_errors.Wrap(err, "ошибка репозитория")
	}
	return err
}
