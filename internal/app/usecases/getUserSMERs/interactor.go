package getUserSMERsUsecase

import (
	"context"

	"github.com/M-Astrid/cbt-helper/internal/app/port"
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
)

type Interactor struct {
	storage port.SMERStorageI
}

func NewInteractor(storage port.SMERStorageI) *Interactor {
	return &Interactor{
		storage: storage,
	}
}

func (i *Interactor) Call(ctx context.Context, userId int64) ([]*entity.SMEREntry, error) {
	return i.storage.GetByUserID(ctx, userId)
}
