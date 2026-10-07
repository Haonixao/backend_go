package format_errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type TracedError struct {
	location string
	message  string
	cause    error
}

func (e *TracedError) Error() string {
	if len(e.message) > 0 {
		return e.message + ": " + e.cause.Error()
	}
	return e.cause.Error()
}

func (e *TracedError) Unwrap() error {
	return e.cause
}

// Wrap оборачивает ошибку и автоматически захватывает файл + строку вызова Wrap
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	// Захватываем место, где вызвали Wrap (skip=1 — пропускаем саму функцию Wrap)
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Делаем путь короче и красивее (убираем /go/src/... или длинный GOPATH)
	shortFile := shortenFilePath(file)

	location := fmt.Sprintf("%s:%d", shortFile, line)

	return &TracedError{
		location: location,
		message:  message,
		cause:    err,
	}
}

// Вспомогательные функции для красивого вывода пути
func shortenFilePath(file string) string {
	// Убираем всё до последнего /cmd/, /internal/, /pkg/ и т.д.
	if idx := strings.LastIndex(file, "/internal/"); idx != -1 {
		return file[idx+1:]
	}
	if idx := strings.LastIndex(file, "/pkg/"); idx != -1 {
		return file[idx+1:]
	}
	if idx := strings.LastIndex(file, "/cmd/"); idx != -1 {
		return file[idx+1:]
	}
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		return file[idx+1:]
	}
	return file
}

func FormatTree(err error) string {
	if err == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("Error trace:\n")

	depth := 0
	current := err

	for current != nil {
		prefix := strings.Repeat("  ", depth)
		if depth > 0 {
			prefix += "└─"
		} else {
			prefix = "└─"
		}

		if te, ok := current.(*TracedError); ok {
			sb.WriteString(fmt.Sprintf("%s%s %s\n", prefix, te.location, te.message))
		} else {
			// Обычная ошибка (из БД, stdlib и т.д.)
			sb.WriteString(fmt.Sprintf("%s%s\n", prefix, current.Error()))
		}

		current = errors.Unwrap(current)
		depth++
	}

	return sb.String()
}
