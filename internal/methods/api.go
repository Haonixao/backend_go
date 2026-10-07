package methods

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"backend_go/internal/models"
	"backend_go/pkg/gorm_extra"

	"backend_go/pkg/format_errors"
)

type fieldError struct {
	err validator.FieldError
}

func (q fieldError) String() string {
	var sb strings.Builder
	fieldName := q.err.Field()
	sb.WriteString("ошибка валидации для поля '" + fieldName + "'")
	sb.WriteString(", условие: " + q.err.ActualTag())
	if q.err.Param() != "" {
		sb.WriteString(" { " + q.err.Param() + " }")
	}
	if q.err.Value() != nil && q.err.Value() != "" {
		sb.WriteString(fmt.Sprintf(", значение: %v", q.err.Value()))
	}
	return sb.String()
}

func ValidationError(err error) error {
	var vErrs validator.ValidationErrors
	if errors.As(err, &vErrs) {
		return format_errors.Wrap(errors.New(fieldError{err: vErrs[0]}.String()), "")
	}
	return err
}

// checkDuplicateKeyViolation
// для проверки нарушения уникальности при создании и обновлении сущностей
func checkDuplicateKeyViolation(err error) (error, bool) {
	var pgE *pgconn.PgError
	if strings.Contains(err.Error(), "SQLSTATE 23505") && errors.As(err, &pgE) {
		err = format_errors.Wrap(errors.New(pgE.Detail), "ошибка дублирования ключа")
		return err, true
	}
	return err, false
}

func DuplicationError(err error) (error, int) {
	if err, ok := checkDuplicateKeyViolation(err); ok {
		return format_errors.Wrap(err, "ошибка checkDuplicateKeyViolation"), http.StatusBadRequest
	}
	return format_errors.Wrap(err, "ошибка checkDuplicateKeyViolation"), http.StatusInternalServerError
}

func GetExistedItems[T any, ID comparable](
	c *gin.Context,
	ids []ID,
	fetcher func([]ID) ([]T, error),
	idGetter func(T) ID,
	entityName string,
) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}
	items, err := fetcher(ids)
	if err != nil {
		errS := fmt.Sprintf("ошибка получения %s по id", entityName)
		c.JSON(http.StatusInternalServerError, models.MessageResponse{Message: errS})
		return nil, format_errors.Wrap(err, errS)
	}
	if len(ids) != len(items) {
		existedSet := make(map[ID]bool)
		for _, item := range items {
			existedSet[idGetter(item)] = true
		}
		var notFound []ID
		for _, id := range ids {
			if !existedSet[id] {
				notFound = append(notFound, id)
			}
		}
		errS := fmt.Sprintf("некоторые %s не найдены по id: %v", entityName, notFound)
		c.JSON(http.StatusNotFound, models.MessageResponse{Message: errS})
		return nil, format_errors.Wrap(errors.New(errS), errS)
	}
	return items, nil
}

func BindAndValidate(c *gin.Context, data interface{}, bindUri bool) error {
	var err error
	if bindUri {
		err = c.ShouldBindUri(data)
	} else {
		err = c.ShouldBind(data)
	}
	if err == nil {
		if v, ok := data.(interface{ Validate() error }); ok {
			err = v.Validate()
		}
	}
	if err != nil {
		err := ValidationError(err)
		return format_errors.Wrap(err, "ошибка привязки и валидации")
	}
	return nil
}

func GetPageParams(c *gin.Context, orderFields ...string) (*gorm_extra.PageParams, error) {
	params := &gorm_extra.PageParams{}
	err := c.ShouldBind(params)
	if err == nil {
		err = params.Validate(orderFields...)
	}
	if err != nil {
		err := ValidationError(err)
		return nil, format_errors.Wrap(err, "ошибка валидации параметров")
	}
	return params, nil
}

func MapPage[T any, U any](page *gorm_extra.Page[T], mapper func(*T) *U) *gorm_extra.Page[U] {
	newPage := &gorm_extra.Page[U]{
		Page:  page.Page,
		Items: make([]*U, len(page.Items)),
	}
	for i, item := range page.Items {
		newPage.Items[i] = mapper(item)
	}
	return newPage
}

func GetUuidFromPath(c *gin.Context, idName string) (uuid.UUID, error) {
	var (
		nId uuid.UUID
		err error
	)
	if id := c.Param(idName); id != "" {
		nId, err = uuid.Parse(id)
	} else {
		err = errors.New(idName + " не задано")
	}
	if err != nil {
		return uuid.Nil, format_errors.Wrap(err, "ошибка получения uuid")
	}
	return nId, nil
}

func ValidateUuids(ids []string) ([]uuid.UUID, error) {
	validIds := []uuid.UUID{}
	for _, id := range ids {
		parsedId, err := uuid.Parse(id)
		if err != nil {
			return nil, format_errors.Wrap(err, "ошибка валидации uuid")
		}
		validIds = append(validIds, parsedId)
	}
	return validIds, nil
}
