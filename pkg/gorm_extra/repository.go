package gorm_extra

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/morkid/paginate"
	"gorm.io/gorm"

	"backend_go/pkg/format_errors"
)

var ChunkSize = 100

type BaseRepository[T any, ID comparable] interface {
	GetConn() (*gorm.DB, error)
	Create(entity *T) error
	CreateWithTx(tx *gorm.DB, entity *T) error
	GetWithTx(tx *gorm.DB) (*T, error)
	GetPage(params *PageParams, preloads ...string) (*Page[T], error)
	GetPageWithTx(tx *gorm.DB, params *PageParams, preloads ...string) (*Page[T], error)
	Get(entity *T, preloads ...string) (*T, error)
	GetMany(entity *T, preloads ...string) ([]*T, error)
	GetManyWithTx(tx *gorm.DB, preloads ...string) ([]*T, error)
	GetAll() ([]*T, error)
	Update(entity *T) error
	UpdateWithTx(tx *gorm.DB, entity *T) error
	Delete(entity *T) error
	DeleteWithTx(tx *gorm.DB) error
	DeleteMany(tx *gorm.DB, entities []*T) error
	CreateManyWithTx(tx *gorm.DB, entities []*T) error
	UpdateManyWithTx(tx *gorm.DB, entities []*T) error
	GetManyByIds(ids []ID) ([]*T, error)
	GetManyByIdsNotIn(ids []ID) ([]*T, error)
	GetExistingIds(ids []ID) ([]ID, error)
}

type BaseRepositoryImpl[T any, ID comparable] struct {
	PCfg *Postgres
	// Другие конфиги для других бд
}

func (s *BaseRepositoryImpl[T, ID]) GetConn() (*gorm.DB, error) {
	// В зависимости от того какой из конфигов не пустой вызвать нужный func (cfg) (*gorm.DB, error)
	conn, err := GetPostgresConn(s.PCfg)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return conn, nil
}

func (s *BaseRepositoryImpl[T, ID]) Create(entity *T) error {
	conn, err := s.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	err = conn.Create(entity).Error
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) CreateWithTx(tx *gorm.DB, entity *T) error {
	err := tx.Create(entity).Error
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) GetWithTx(tx *gorm.DB) (*T, error) {
	var resEntity *T
	if err := tx.First(&resEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return resEntity, nil
}

func (s *BaseRepositoryImpl[T, ID]) GetPage(params *PageParams, preloads ...string) (*Page[T], error) {
	query, err := s.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	query = query.Model(new(T))
	res, err := s.GetPageWithTx(query, params, preloads...)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return res, nil
}

func AddRelevanceOrders(tx *gorm.DB, k string, rV string) *gorm.DB {
	if rV == "" || k == "" {
		return tx
	}

	// Экранируем спецсимволы регулярных выражений и одинарные кавычки (для postgres)
	rV = strings.NewReplacer(
		`\`, `\\`, `^`, `\^`, `$`, `\$`, `.`, `\.`, `|`, `\|`,
		`?`, `\?`, `*`, `\*`, `+`, `\+`, `(`, `\(`, `)`, `\)`,
		`[`, `\[`, `]`, `\]`, `{`, `\{`, `}`, `\}`,
		`'`, `''`,
	).Replace(rV)

	return tx.
		// строгое начало > строгая середина > строгий конец
		Order(fmt.Sprintf(`(lower(("%s")::text) ~* '^%s(\s|$)') DESC`, k, rV)).
		Order(fmt.Sprintf(`(lower(("%s")::text) ~* '\s%s\s') DESC`, k, rV)).
		Order(fmt.Sprintf(`(lower(("%s")::text) ~* '%s$') DESC`, k, rV)).
		// > нестрогое начало > нестрогая середина > нестрогий конец
		Order(fmt.Sprintf(`(lower(("%s")::text) ~* '^[^\s]*%s[^\s]*(\s|$)') DESC`, k, rV)).
		Order(fmt.Sprintf(`(lower(("%s")::text) ~* '\s[^\s]*%s[^\s]*\s') DESC`, k, rV)).
		Order(fmt.Sprintf(`(lower(("%s")::text) ~* '[^\s]*%s[^\s]*$') DESC`, k, rV)).
		// > любое оставшееся место
		Order(fmt.Sprintf(`(lower(("%s")::text) ~* '.*%s.*') DESC`, k, rV))
}

func GetPageWithTx[T any](tx *gorm.DB, params *PageParams, preloads ...string) (*Page[T], error) {
	if tx.Error != nil {
		return nil, format_errors.Wrap(tx.Error, "исходный запрос содержит ошибку")
	}
	type relevanceOrder struct {
		k string
		v string
	}
	var relevanceOrders []relevanceOrder
	req := http.Request{
		URL: &url.URL{},
	}
	reqValues := req.URL.Query()

	filtersOperator := "or"
	if params.FiltersAndOrSwitch != nil && *params.FiltersAndOrSwitch == "and" {
		filtersOperator = "and"
	}

	filter := ""
	for k, v := range params.Filters {
		if v != "" {
			// Сначала экранируем обратный слеш, затем двойные кавычки для JSON
			safeV := strings.ReplaceAll(v, `\`, `\\`)
			safeV = strings.ReplaceAll(safeV, `"`, `\"`)
			filter += fmt.Sprintf(`["%s"],["%s","like","%s"],`, filtersOperator, k, safeV)
			relevanceOrders = append(relevanceOrders, relevanceOrder{k: k, v: v})
		}
	}

	filter = "[" + strings.TrimSuffix(strings.TrimPrefix(filter, fmt.Sprintf(`["%s"],`, filtersOperator)), ",") + "]"

	if filter != "[]" {
		oldPreloads := tx.Statement.Preloads
		isUnscoped := tx.Statement.Unscoped
		reqValues.Set("filters", filter)
		outerTx := tx.Session(&gorm.Session{NewDB: true}).Table("(?) as qro", tx).Model(new(T))
		for _, ro := range relevanceOrders {
			outerTx = AddRelevanceOrders(outerTx, ro.k, ro.v)
		}
		if isUnscoped {
			outerTx = outerTx.Unscoped()
		}
		outerTx.Statement.Preloads = oldPreloads
		tx = outerTx
	}

	if params.Page != nil {
		reqValues.Set("page", strconv.Itoa(*params.Page))
	}
	if params.Size != nil {
		reqValues.Set("size", strconv.Itoa(*params.Size))
	}
	if params.Sort != nil {
		reqValues.Set("sort", *params.Sort)
	}

	for _, p := range preloads {
		if p != "" {
			tx = tx.Preload(p)
		}
	}

	req.URL.RawQuery = reqValues.Encode()
	res := paginate.New(&paginate.Config{
		PageStart: 1,
	}).With(tx).Request(req).Response(&[]*T{})
	if res.RawError != nil {
		return nil, format_errors.Wrap(res.RawError, "ошибка получения страницы")
	}
	resPage := &Page[T]{
		Page: res,
	}
	resPage.Items = *res.Items.(*[]*T)
	return resPage, nil
}

func (s *BaseRepositoryImpl[T, ID]) GetPageWithTx(tx *gorm.DB, params *PageParams, preloads ...string) (*Page[T], error) {
	res, err := GetPageWithTx[T](tx, params, preloads...)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return res, nil
}

func (s *BaseRepositoryImpl[T, ID]) Get(entity *T, preloads ...string) (*T, error) {
	conn, err := s.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	var resEntity *T
	query := conn
	for _, preload := range preloads {
		if preload != "" {
			query = query.Preload(preload)
		}
	}
	if err := query.Where(entity).First(&resEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return resEntity, nil
}

func (s *BaseRepositoryImpl[T, ID]) GetMany(entity *T, preloads ...string) ([]*T, error) {
	conn, err := s.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	query := conn
	for _, preload := range preloads {
		if preload != "" {
			query = query.Preload(preload)
		}
	}
	var entityList []*T
	if err := query.Where(entity).Find(&entityList).Error; err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return entityList, nil
}

func (s *BaseRepositoryImpl[T, ID]) GetManyWithTx(tx *gorm.DB, preloads ...string) ([]*T, error) {
	var entityList []*T
	for _, preload := range preloads {
		if preload != "" {
			tx = tx.Preload(preload)
		}
	}
	if err := tx.Find(&entityList).Error; err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return entityList, nil
}

func (s *BaseRepositoryImpl[T, ID]) GetAll() ([]*T, error) {
	conn, err := s.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	var entities []*T
	if err := conn.Find(&entities).Error; err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return entities, nil
}

func (s *BaseRepositoryImpl[T, ID]) Update(entity *T) error {
	conn, err := s.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	res := conn.Omit("CreatedAt").Save(entity)
	if res.Error != nil {
		return format_errors.Wrap(res.Error, "ошибка репозитория")
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) UpdateWithTx(tx *gorm.DB, entity *T) error {
	err := tx.Omit("CreatedAt").Save(entity).Error
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) Delete(entity *T) error {
	conn, err := s.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	res := conn.Delete(&entity)
	if res.Error != nil {
		return format_errors.Wrap(res.Error, "ошибка репозитория")
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) DeleteWithTx(tx *gorm.DB) error {
	err := tx.Delete(new(T)).Error
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) DeleteMany(tx *gorm.DB, entities []*T) error {
	for i := 0; i <= len(entities)/ChunkSize; i++ {
		end := (i + 1) * ChunkSize
		if end > len(entities) {
			end = len(entities)
		}
		entitiesChunk := entities[i*ChunkSize : end]
		if len(entitiesChunk) > 0 {
			if err := tx.Delete(&entitiesChunk).Error; err != nil {
				return format_errors.Wrap(err, "ошибка репозитория")
			}
		}
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) CreateManyWithTx(tx *gorm.DB, entities []*T) error {
	err := tx.CreateInBatches(entities, ChunkSize).Error
	if err != nil {
		return format_errors.Wrap(err, "ошибка репозитория")
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) UpdateManyWithTx(tx *gorm.DB, entities []*T) error {
	for _, entity := range entities {
		if err := tx.Omit("CreatedAt").Save(entity).Error; err != nil {
			return format_errors.Wrap(err, "ошибка репозитория")
		}
	}
	return nil
}

func (s *BaseRepositoryImpl[T, ID]) GetManyByIds(ids []ID) ([]*T, error) {
	conn, err := s.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	query := conn.Where("id in ?", ids)
	res, err := s.GetManyWithTx(query)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return res, nil
}

func (s *BaseRepositoryImpl[T, ID]) GetManyByIdsNotIn(ids []ID) ([]*T, error) {
	conn, err := s.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	query := conn.Where("id not in ?", ids)
	res, err := s.GetManyWithTx(query)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return res, nil
}

func (s *BaseRepositoryImpl[T, ID]) GetExistingIds(ids []ID) ([]ID, error) {
	conn, err := s.GetConn()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	var existingIds []ID
	if err := conn.Model(new(T)).Where("id IN ?", ids).Pluck("id", &existingIds).Error; err != nil {
		return nil, format_errors.Wrap(err, "ошибка репозитория")
	}
	return existingIds, nil
}
