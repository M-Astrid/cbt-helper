package deleteSMERUsecase

import (
	"context"

	"github.com/M-Astrid/cbt-helper/internal/app/port"
)

type Interactor struct {
	storage port.SMERStorageI
}

func NewInteractor(storage port.SMERStorageI) *Interactor {
	return &Interactor{
		storage: storage,
	}
}

func (i *Interactor) Call(ctx context.Context, id string) error {
	return i.storage.DeleteByID(ctx, id)
}
