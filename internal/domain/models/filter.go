package models

import (
	domainerrors "github.com/maksemen2/pvz-service/internal/domain/errors"
	"time"
)

// PVZFilter - структура для инкапсуляции фильтров для вывода ПВЗ.
// Опциональны только поля StartDate и EndDate.
// Page и PageSize должны подставляться на уровне бизнес логики
type PVZFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
}

// Valid проводит валидацию PVZFilter. Возвращает доменные ошибки.
func (f *PVZFilter) Valid() error {
	if f.Page < 1 {
		return domainerrors.ErrInvalidPage
	}

	if f.PageSize < 1 {
		return domainerrors.ErrInvalidLimit
	}

	// Лучше сразу проверить фильтры по дате, чтобы не делать лишней нагрузки на
	// Заведомо пустые ответы
	if f.StartDate != nil && f.EndDate != nil && f.StartDate.After(*f.EndDate) {
		return domainerrors.ErrInvalidDateRange
	}

	if f.StartDate != nil && f.StartDate.After(time.Now()) {
		return domainerrors.ErrInvalidStartDate
	}

	return nil
}
