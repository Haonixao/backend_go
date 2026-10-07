package backup_services

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"backend_go/pkg/format_errors"

	"backend_go/internal/models/public"
	"backend_go/internal/repositories/backup_repositories"
	"backend_go/pkg/gorm_extra"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Стратегии восстановления
const (
	RestoreStrategySkip    = "skip"    // Пропускать существующие записи (ON CONFLICT DO NOTHING)
	RestoreStrategyReplace = "replace" // Перезаписывать существующие записи (ON CONFLICT DO UPDATE)
	RestoreStrategyFail    = "fail"    // Прерывать при конфликте (без ON CONFLICT)
)

type BackupService interface {
	gorm_extra.BaseService[public.Backup, uuid.UUID]
	CreateBackup() (*public.Backup, error)
	RestoreBackup(backupId uuid.UUID, strategy string) error
	DeleteOldBackups(beforeDate time.Time) error
}

type BackupServiceImpl struct {
	gorm_extra.BaseService[public.Backup, uuid.UUID]
	BackupRep backup_repositories.BackupRepository
}

func (s *BackupServiceImpl) CreateBackup() (*public.Backup, error) {
	var migrationVersion string
	conn, err := s.BackupRep.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	err = conn.Raw(`
		SELECT id::text
		FROM public.migrations
		ORDER BY applied_at DESC
		LIMIT 1
	`).Scan(&migrationVersion).Error
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	if migrationVersion == "" {
		err := errors.New("ни одной версии миграции еще не создано")
		return nil, err
	}
	var jsonData string
	// 'roles', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.role t),
	// 'permissions', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.permission t),
	// 'permission2role', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.permission2role t),
	// 'regions', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM iod_data.region t),
	// 'organization_types', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM iod_data.organization_type t),
	// 'organizations', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM iod_data.organization t),
	// 'persons', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM iod_data.person t),
	// 'users', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.user t),
	// 'user2role', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.user2role t),
	// 'user_history', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.user_history t),
	// 'applications', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.application t),
	// 'application2role', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.application2role t),
	// 'application_domains', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.application_domain t),
	// 'mail_hosts', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.mail_host t),
	// 'admin_settings', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.admin_settings t),
	// 'auth_host_accounts', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.auth_host_account t),
	// 'invalid_password_chars', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM sea.invalid_password_chars t)
	err = conn.Raw(`
		SELECT json_build_object(
			'tasks', (SELECT COALESCE(json_agg(t.*), '[]'::json) FROM scheduled.task t)
		)::text
	`).Scan(&jsonData).Error
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	compressed, err := CompressData([]byte(jsonData))
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	sizeMB := float64(len(compressed)) / 1024.0 / 1024.0
	sizeStr := fmt.Sprintf("%.2fМб", sizeMB)
	backupModel := &public.Backup{
		Size:             sizeStr,
		MigrationVersion: migrationVersion,
		Data:             compressed,
	}
	err = s.BackupRep.Create(backupModel)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}

	return backupModel, nil
}

// getWhereFilterForTable возвращает WHERE условие для фильтрации битых FK ссылок
func getWhereFilterForTable(tableName string) string {
	switch tableName {
	// case "user2role":
	// 	return `WHERE EXISTS (SELECT 1 FROM sea.user WHERE id = t.user_id)
	// 		AND EXISTS (SELECT 1 FROM sea.role WHERE id = t.role_id)`
	// case "permission2role":
	// 	return `WHERE EXISTS (SELECT 1 FROM sea.permission WHERE composite_id = t.permission_composite_id)
	// 		AND EXISTS (SELECT 1 FROM sea.role WHERE id = t.role_id)`
	// case "application2role":
	// 	return `WHERE EXISTS (SELECT 1 FROM sea.application WHERE id = t.application_id)
	// 		AND EXISTS (SELECT 1 FROM sea.role WHERE id = t.role_id)`
	// case "application_domains":
	// 	return `WHERE EXISTS (SELECT 1 FROM sea.application WHERE id = t.application_id)`
	// case "auth_host_accounts":
	// 	return `WHERE EXISTS (SELECT 1 FROM sea.user WHERE id = t.user_id)`
	// case "user_history":
	// 	return `WHERE t.admin_user_id IS NULL OR EXISTS (SELECT 1 FROM sea.user WHERE id = t.admin_user_id)`
	default:
		return ""
	}
}

func (s *BackupServiceImpl) RestoreBackup(backupId uuid.UUID, strategy string) error {
	if strategy != RestoreStrategySkip && strategy != RestoreStrategyReplace && strategy != RestoreStrategyFail {
		err := fmt.Errorf("некорректная стратегия восстановления: %s. Доступные: %s, %s, %s",
			strategy, RestoreStrategySkip, RestoreStrategyReplace, RestoreStrategyFail)
		return err
	}
	backupModel, err := s.BackupRep.Get(&public.Backup{
		UuidId: gorm_extra.UuidId{ID: backupId},
	})
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	if backupModel == nil {
		err := errors.New("бэкап не найден")
		return err
	}
	jsonData, err := DecompressData(backupModel.Data)
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}
	conn, err := s.BackupRep.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	err = conn.Transaction(func(tx *gorm.DB) error {
		tableOrder := []struct {
			name      string
			tableName string
		}{
			{"tasks", "scheduled.task"},
			// {"roles", "sea.role"},
			// {"applications", "sea.application"},
			// {"regions", "iod_data.region"},
			// {"organization_types", "iod_data.organization_type"},
			// {"organizations", "iod_data.organization"},
			// {"persons", "iod_data.person"},
			// {"users", "sea.user"},
			// {"mail_hosts", "sea.mail_host"},
			// {"admin_settings", "sea.admin_settings"},
			// {"invalid_password_chars", "sea.invalid_password_chars"},
			// {"permissions", "sea.permission"},
			// {"permission2role", "sea.permission2role"},
			// {"user2role", "sea.user2role"},
			// {"application2role", "sea.application2role"},
			// {"application_domains", "sea.application_domain"},
			// {"auth_host_accounts", "sea.auth_host_account"},
			// {"user_history", "sea.user_history"},
		}
		for _, table := range tableOrder {
			tableData, exists := data[table.name]
			if !exists || len(tableData) == 0 {
				continue
			}
			whereFilter := getWhereFilterForTable(table.name)
			var err error
			switch strategy {
			case RestoreStrategySkip:
				query := fmt.Sprintf(`
					INSERT INTO %s
					SELECT t.* FROM json_populate_recordset(NULL::%s, $1::json) t
					%s
					ON CONFLICT DO NOTHING
				`, table.tableName, table.tableName, whereFilter)
				err = tx.Exec(query, string(tableData)).Error

			case RestoreStrategyReplace:
				err = tx.Exec(fmt.Sprintf(`TRUNCATE TABLE %s CASCADE`, table.tableName)).Error
				if err != nil {
					return format_errors.Wrap(err, "ошибка сервиса")
				}
				query := fmt.Sprintf(`
					INSERT INTO %s
					SELECT t.* FROM json_populate_recordset(NULL::%s, $1::json) t
					%s
				`, table.tableName, table.tableName, whereFilter)
				err = tx.Exec(query, string(tableData)).Error

			case RestoreStrategyFail:
				query := fmt.Sprintf(`
					INSERT INTO %s
					SELECT t.* FROM json_populate_recordset(NULL::%s, $1::json) t
					%s
				`, table.tableName, table.tableName, whereFilter)
				err = tx.Exec(query, string(tableData)).Error
			}
			if err != nil {
				return format_errors.Wrap(err, "ошибка сервиса")
			}
		}
		return nil
	})
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return nil
}

func (s *BackupServiceImpl) DeleteOldBackups(beforeDate time.Time) error {
	err := s.BackupRep.DeleteOldBackups(beforeDate)
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return nil
}

func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecompressData(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	defer gzipReader.Close()
	return io.ReadAll(gzipReader)
}
