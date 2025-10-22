package getUserSMERsUsecase

import (
	"context"
	"time"

	"github.com/M-Astrid/cbt-helper/internal/app/port"
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
)

type Interactor struct {
	storage port.SMERStorage
}

func NewInteractor(storage port.SMERStorage) *Interactor {
	return &Interactor{
		storage: storage,
	}
}

func (i *Interactor) Call(ctx context.Context, userId int64, startDate time.Time, endDate time.Time) ([]*entity.SMEREntry, error) {
	return i.storage.GetByUserID(ctx, userId, startDate, endDate)
}
