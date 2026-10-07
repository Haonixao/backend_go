package status

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"backend_go/internal/config"
	"backend_go/internal/methods"
	"backend_go/internal/models"
	"backend_go/internal/models/public"
	"backend_go/internal/models/scheduled_models"
	"backend_go/internal/services/backup_services"
	"backend_go/pkg/format_errors"
	"backend_go/pkg/gorm_extra"
	"backend_go/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Handler struct {
	BackupService    backup_services.BackupService
	TaskLocksService gorm_extra.BaseService[scheduled_models.TaskLock, string]
	Logger           *zerolog.Logger
}

// GetStatus
//
//	@Tags		status
//	@Produce	json
//	@Description	Статус работоспособности
//	@Response	200	{object}	ResponseStatus
//	@Failure	401	{object}	models.MessageResponse
//	@Failure	403	{object}	models.MessageResponse
//	@Router		/api/v1/status/check [get]
func (h *Handler) GetStatus(c *gin.Context) {
	db, err := gorm_extra.GetPostgresConn(config.GetConfig().Postgres)
	status := ResponseStatus{
		Service: "Работает",
	}
	if err == nil {
		var value int
		db.Raw("SELECT 1").Scan(&value)
		if value != 1 {
			status.Database = "Не работает"
		} else {
			status.Database = "Работает"
		}
	} else {
		status.Database = "Не работает"
		h.Logger.Error().Msg(format_errors.FormatTree(err))
	}
	c.JSON(http.StatusOK, status)
}

// GetVersions
//
//	@Tags		status
//	@Produce	json
//	@Description Список версий структуры данных
//	@Success	200	{array}		Version
//	@Failure	500	{object}	models.MessageResponse
//	@Failure	401	{object}	models.MessageResponse
//	@Failure	403	{object}	models.MessageResponse
//	@Router		/api/v1/status/versions [get]
func (h *Handler) GetVersions(c *gin.Context) {
	db, err := gorm_extra.GetPostgresConn(config.GetConfig().Postgres)
	if err != nil {
		errS := "ошибка подключения к бд: " + err.Error()
		h.Logger.Error().Msg(format_errors.FormatTree(err))
		c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
		return
	}
	var versions []*Version
	db.Raw("SELECT id, applied_at FROM public.migrations ORDER BY applied_at DESC").Scan(&versions)
	if versions == nil {
		errS := "версии не найдены. что-то не так с бд"
		h.Logger.Error().Msg(errS)
		c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
		return
	}
	c.JSON(http.StatusOK, versions)
}

// GetBackups
//
//	@Tags		status
//	@Produce	json
//	@Description	Получить список бэкапов
//	@Param		pageParams	query		gorm_extra.PageParams	true	"параметры пагинации"
//	@Success	200	{object}	Page[BackupDTO]
//	@Failure	400	{object}	models.MessageResponse
//	@Failure	401	{object}	models.MessageResponse
//	@Failure	403	{object}	models.MessageResponse
//	@Failure	500	{object}	models.MessageResponse
//	@Router		/api/v1/status/backups [get]
func (h *Handler) GetBackups(c *gin.Context) {
	params, err := methods.GetPageParams(c, "created_at")
	if err != nil {
		errS := "ошибка валидации: " + err.Error()
		h.Logger.Error().Msg(format_errors.FormatTree(err))
		c.JSON(http.StatusBadRequest, models.MessageResponse{Message: errS})
		return
	}
	backupsPage, err := h.BackupService.GetPage(params)
	if err != nil {
		errS := "ошибка получения бэкапов: " + err.Error()
		h.Logger.Error().Msg(format_errors.FormatTree(err))
		c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
		return
	}
	backupDTOs := make([]*BackupDTO, len(backupsPage.Items))
	for i, backup := range backupsPage.Items {
		backupDTOs[i] = &BackupDTO{
			ID:               backup.ID,
			CreatedAt:        backup.CreatedAt,
			Size:             backup.Size,
			MigrationVersion: backup.MigrationVersion,
		}
	}
	response := gorm_extra.Page[BackupDTO]{
		Page:  backupsPage.Page,
		Items: backupDTOs,
	}
	c.JSON(http.StatusOK, response)
}

// GetLogs
//
// @Tags		status
// @Produce	json
// @Description Получить логи
// @Failure	500	{object}	models.MessageResponse
// @Failure	401	{object}	models.MessageResponse
// @Failure	403	{object}	models.MessageResponse
// @Success		200	{file}		string	"Plain text log file"
// @Router		/api/v1/status/logs [get]
func (h *Handler) GetLogs(dir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/plain")
		c.Header("Content-Disposition", "attachment; filename=\"app.log\"")

		err := log.StreamLogs(c.Writer, dir)
		if err != nil {
			errS := "ошибка получения логов: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
	}
}

// ApplyBackup
//
//	@Tags		status
//	@Accept		json
//	@Produce	json
//	@Description	Применить бэкап данных
//	@Param		id		path		string					true	"ID бэкапа"
//	@Param		body	body		ApplyBackupRequest		true	"Стратегия восстановления (skip, replace, fail)"
//	@Success	200	{object}	models.MessageResponse
//	@Failure	400	{object}	models.MessageResponse
//	@Failure	401	{object}	models.MessageResponse
//	@Failure	403	{object}	models.MessageResponse
//	@Failure	404	{object}	models.MessageResponse
//	@Failure	409	{object}	models.MessageResponse
//	@Failure	500	{object}	models.MessageResponse
//	@Router		/api/v1/status/backups/{id}/apply [post]
func (h *Handler) ApplyBackup(isReady *atomic.Bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid, err := methods.GetUuidFromPath(c, "id")
		if err != nil {
			errS := "ошибка получения id из пути: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusBadRequest, models.MessageResponse{Message: errS})
			return
		}
		var bodyP ApplyBackupRequest
		if err := methods.BindAndValidate(c, &bodyP, false); err != nil {
			errS := "ошибка валидации: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusBadRequest, models.MessageResponse{Message: errS})
			return
		}
		backup := &public.Backup{UuidId: gorm_extra.UuidId{ID: uuid}}
		backup, err = h.BackupService.Get(backup)
		if err != nil {
			errS := "ошибка получения бэкапа: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		if backup == nil {
			errS := "бэкап не найден"
			h.Logger.Error().Msg(errS)
			c.JSON(http.StatusNotFound, models.MessageResponse{Message: errS})
			return
		}
		backupTaskLock, err := h.TaskLocksService.Get(&scheduled_models.TaskLock{TaskID: "create_periodic_backup"})
		if err != nil {
			errS := "ошибка проверки блокировки: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		if backupTaskLock != nil {
			errS := "невозможно применить бэкап: в данный момент создается новый бэкап или применяется другой"
			h.Logger.Error().Msg(errS)
			c.JSON(http.StatusConflict, models.MessageResponse{Message: errS})
			return
		}
		isReady.Store(false)
		defer isReady.Store(true)
		backupTaskLock = &scheduled_models.TaskLock{
			TaskID:    "create_periodic_backup",
			CreatedAt: time.Now(),
		}
		err = h.TaskLocksService.Create(backupTaskLock)
		if err != nil {
			errS := "ошибка создания блокировки: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		defer func() {
			err = h.TaskLocksService.Delete(backupTaskLock)
			if err != nil {
				h.Logger.Error().Msg(format_errors.FormatTree(err))
			}
		}()
		h.Logger.Info().
			Str("backup_id", uuid.String()).
			Str("strategy", bodyP.Strategy).
			Msg("Начато применение бэкапа")
		err = h.BackupService.RestoreBackup(uuid, bodyP.Strategy)
		if err != nil {
			errS := "ошибка применения бэкапа: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		h.Logger.Info().
			Str("backup_id", uuid.String()).
			Str("strategy", bodyP.Strategy).
			Msg("Бэкап успешно применен")
		c.JSON(http.StatusOK, models.MessageResponse{Message: "Бэкап успешно применен"})
	}
}

// GetBackupData
//
//	@Tags		status
//	@Produce	json
//	@Description	Получить сохраненные данные бэкапа
//	@Param		id	path		string	true	"id бэкапа"
//	@Success	200	{object}	map[string]any
//	@Failure	400	{object}	models.MessageResponse
//	@Failure	401	{object}	models.MessageResponse
//	@Failure	403	{object}	models.MessageResponse
//	@Failure	404	{object}	models.MessageResponse
//	@Failure	500	{object}	models.MessageResponse
//	@Router		/api/v1/status/backups/{id}/data [get]
func (h *Handler) GetBackupData() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid, err := methods.GetUuidFromPath(c, "id")
		if err != nil {
			errS := "ошибка получения id из пути: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusBadRequest, models.MessageResponse{Message: errS})
			return
		}
		backup := &public.Backup{UuidId: gorm_extra.UuidId{ID: uuid}}
		backup, err = h.BackupService.Get(backup)
		if err != nil {
			errS := "ошибка получения бэкапа: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		if backup == nil {
			errS := "бэкап не найден"
			h.Logger.Error().Msg(errS)
			c.JSON(http.StatusNotFound, models.MessageResponse{Message: errS})
			return
		}
		jsonData, err := backup_services.DecompressData(backup.Data)
		if err != nil {
			errS := "ошибка получения данных бэкапа: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		var data map[string]json.RawMessage
		if err := json.Unmarshal(jsonData, &data); err != nil {
			errS := "ошибка получения данных бэкапа: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		c.JSON(http.StatusOK, data)
	}
}

// CreateBackup
//
//	@Tags		status
//	@Produce	json
//	@Description	Создать бэкап
//	@Success	200	{object}	BackupDTO
//	@Failure	401	{object}	models.MessageResponse
//	@Failure	403	{object}	models.MessageResponse
//	@Failure	409	{object}	models.MessageResponse
//	@Failure	500	{object}	models.MessageResponse
//	@Router		/api/v1/status/backups [post]
func (h *Handler) CreateBackup(isReady *atomic.Bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		backupTaskLock, err := h.TaskLocksService.Get(&scheduled_models.TaskLock{TaskID: "create_periodic_backup"})
		if err != nil {
			errS := "ошибка проверки блокировки: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		if backupTaskLock != nil {
			errS := "невозможно создать бэкап: в данный момент создается другой бэкап или применяется существующий"
			h.Logger.Error().Msg(errS)
			c.JSON(http.StatusConflict, models.MessageResponse{Message: errS})
			return
		}

		isReady.Store(false)
		defer isReady.Store(true)

		backupTaskLock = &scheduled_models.TaskLock{
			TaskID:    "create_periodic_backup",
			CreatedAt: time.Now(),
		}
		err = h.TaskLocksService.Create(backupTaskLock)
		if err != nil {
			errS := "ошибка создания блокировки: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}
		defer func() {
			err = h.TaskLocksService.Delete(backupTaskLock)
			if err != nil {
				h.Logger.Error().Msg(format_errors.FormatTree(err))
			}
		}()

		backup, err := h.BackupService.CreateBackup()
		if err != nil {
			errS := "ошибка создания бэкапа: " + err.Error()
			h.Logger.Error().Msg(format_errors.FormatTree(err))
			c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
			return
		}

		response := &BackupDTO{
			ID:               backup.ID,
			CreatedAt:        backup.CreatedAt,
			Size:             backup.Size,
			MigrationVersion: backup.MigrationVersion,
		}

		c.JSON(http.StatusOK, response)
	}
}
