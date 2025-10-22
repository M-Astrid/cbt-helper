package getSingleSMERUsecase

import (
	"context"

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

func (i *Interactor) Call(ctx context.Context, id string) (*entity.SMEREntry, error) {
	return i.storage.GetByID(ctx, id)
}
