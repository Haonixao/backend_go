package gorm_extra

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"backend_go/pkg/format_errors"

	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

func getAvailableIdQuery(idColumn string, table string) string {
	return fmt.Sprintf(`
	SELECT COALESCE(
		(
			SELECT MIN(n)
			FROM generate_series(1, (SELECT COALESCE(MAX(%s), 0) FROM %s) + 1) n
			WHERE n NOT IN (SELECT %s FROM %s)
			AND n NOT IN (SELECT unnest($1::bigint[]))
		),
		(
			SELECT GREATEST(0, (SELECT MAX(u)::bigint FROM unnest($2::bigint[]) u)) + 1
		)
	)`,
		idColumn, table, idColumn, table)
}

// GetAvailableId
// Возвращает минимальный доступный ID для таблицы
// usedIds нужен для фиксации использованных id в случае транзакции
func GetAvailableId(tx *gorm.DB, table string) (int64, error) {
	var newID int64
	query := getAvailableIdQuery("id", table)
	checkUsedIds := tx.Statement.Context.Value("usedIds")
	usedIds := []int64{0}
	if checkUsedIds != nil {
		usedIds = checkUsedIds.([]int64)
	}
	err := tx.Raw(query, usedIds, usedIds).Scan(&newID).Error
	if err != nil {
		return 0, format_errors.Wrap(err, "ошибка получения свободного id")
	}
	usedIds = append(usedIds, newID)
	tx.Statement.Context = context.WithValue(tx.Statement.Context, "usedIds", usedIds)
	return newID, nil
}

type Timestamps struct {
	CreatedAt time.Time `gorm:"type:timestamp with time zone;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone;not null" json:"updated_at"`
}

type StrId struct {
	ID string `gorm:"primaryKey;type:text;not null" json:"id"`
}

// SerialId
// Без autoIncrement
// Использовать вместе с GetAvailableId в BeforeCreate модели
type SerialId struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;not null;type:bigint" json:"id"`
}

type UuidId struct {
	ID uuid.UUID `gorm:"type:uuid;not null;primary_key" json:"id"`
}

func (b *UuidId) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}

type Page[T any] struct {
	paginate.Page
	Error        *bool   `json:"-"`
	ErrorMessage *string `json:"-"`
	Items        []*T    `json:"items"`
}

type PageParams struct {
	Page               *int              `form:"page" binding:"omitempty,numeric,min=1" default:"1"`
	Size               *int              `form:"size" binding:"omitempty,numeric,min=1" default:"10"`
	Sort               *string           `form:"sort" example:"'-name,id' --> name desc and id asc"`
	Filters            map[string]string `form:"-"`
	FiltersAndOrSwitch *string           `form:"filters_and_or_switch" example:"'or' || 'and'"`
}

func (p *PageParams) Validate(orderFields ...string) error {
	if p.Page != nil && *p.Page <= 0 {
		errS := "страница должна быть > 0"
		return format_errors.Wrap(errors.New(errS), "")
	}

	if p.Size != nil && (*p.Size <= 0) {
		errS := "размер страницы должен быть > 0"
		return format_errors.Wrap(errors.New(errS), "")
	}
	if p.Sort != nil && len(*p.Sort) != 0 {
		sorts := strings.Split(*p.Sort, ",")
		for _, e := range sorts {
			sortOne, _ := strings.CutPrefix(e, "-")
			if !slices.Contains(orderFields, sortOne) {
				errS := fmt.Sprintf("неправильное поле сортировки: %s. доступные поля: %s", sortOne, orderFields)
				return format_errors.Wrap(errors.New(errS), "")
			}
		}
	}
	return nil
}
