package gorm_extra

import (
	"gorm.io/gorm"

	"backend_go/pkg/format_errors"
)

type BaseService[T any, ID comparable] interface {
	GetAll() ([]*T, error)
	Create(element *T) error
	Update(element *T) error
	CreateMany(elements []*T) error
	GetPage(params *PageParams, preloads ...string) (*Page[T], error)
	Get(element *T, preloads ...string) (*T, error)
	GetMany(entity *T, preloads ...string) ([]*T, error)
	GetManyWithTx(tx *gorm.DB, preloads ...string) ([]*T, error)
	Delete(element *T) error
	DeleteMany(elements []*T) error
	UpdateMany(elements []*T) error
	GetManyByIds(ids []ID) ([]*T, error)
	GetManyByIdsNotIn(ids []ID) ([]*T, error)
	GetExistingIds(ids []ID) ([]ID, error)
}

type BaseServiceImpl[T any, ID comparable] struct {
	BaseRepository BaseRepository[T, ID]
}

func (s *BaseServiceImpl[T, ID]) GetAll() ([]*T, error) {
	elements, err := s.BaseRepository.GetAll()
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	return elements, nil
}

func (s *BaseServiceImpl[T, ID]) GetMany(entity *T, preloads ...string) ([]*T, error) {
	res, err := s.BaseRepository.GetMany(entity, preloads...)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	return res, nil
}

func (s *BaseServiceImpl[T, ID]) GetManyWithTx(tx *gorm.DB, preloads ...string) ([]*T, error) {
	res, err := s.BaseRepository.GetManyWithTx(tx, preloads...)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	return res, nil
}

func (s *BaseServiceImpl[T, ID]) Get(element *T, preloads ...string) (*T, error) {
	res, err := s.BaseRepository.Get(element, preloads...)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	return res, nil
}

func (s *BaseServiceImpl[T, ID]) Create(element *T) error {
	err := s.BaseRepository.Create(element)
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return nil
}

func (s *BaseServiceImpl[T, ID]) Update(element *T) error {
	err := s.BaseRepository.Update(element)
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return nil
}

func (s *BaseServiceImpl[T, ID]) CreateMany(elements []*T) error {
	conn, err := s.BaseRepository.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return conn.Transaction(func(tx *gorm.DB) error {
		err := s.BaseRepository.CreateManyWithTx(tx, elements)
		if err != nil {
			return format_errors.Wrap(err, "ошибка сервиса")
		}
		return nil
	})
}

func (s *BaseServiceImpl[T, ID]) GetPage(params *PageParams, preloads ...string) (*Page[T], error) {
	page, err := s.BaseRepository.GetPage(params, preloads...)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	return page, nil
}

func (s *BaseServiceImpl[T, ID]) Delete(element *T) error {
	err := s.BaseRepository.Delete(element)
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return nil
}

func (s *BaseServiceImpl[T, ID]) DeleteMany(elements []*T) error {
	conn, err := s.BaseRepository.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	err = conn.Transaction(func(tx *gorm.DB) error {
		err := s.BaseRepository.DeleteMany(tx, elements)
		if err != nil {
			return format_errors.Wrap(err, "ошибка сервиса")
		}
		return nil
	})
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return nil
}

func (s *BaseServiceImpl[T, ID]) UpdateMany(elements []*T) error {
	conn, err := s.BaseRepository.GetConn()
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	err = conn.Transaction(func(tx *gorm.DB) error {
		err := s.BaseRepository.UpdateManyWithTx(tx, elements)
		if err != nil {
			return format_errors.Wrap(err, "ошибка сервиса")
		}
		return nil
	})
	if err != nil {
		return format_errors.Wrap(err, "ошибка сервиса")
	}
	return nil
}

func (s *BaseServiceImpl[T, ID]) GetManyByIds(ids []ID) ([]*T, error) {
	res, err := s.BaseRepository.GetManyByIds(ids)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	return res, nil
}

func (s *BaseServiceImpl[T, ID]) GetManyByIdsNotIn(ids []ID) ([]*T, error) {
	res, err := s.BaseRepository.GetManyByIdsNotIn(ids)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	return res, nil
}

func (s *BaseServiceImpl[T, ID]) GetExistingIds(ids []ID) ([]ID, error) {
	res, err := s.BaseRepository.GetExistingIds(ids)
	if err != nil {
		return nil, format_errors.Wrap(err, "ошибка сервиса")
	}
	return res, nil
}
