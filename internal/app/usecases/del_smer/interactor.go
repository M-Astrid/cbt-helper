package delete_smer_usecase

import (
	"context"

	"github.com/M-Astrid/cbt-helper/internal/app/port"
)

type Interactor struct {
	storage port.SMERStorage
}

func NewInteractor(storage port.SMERStorage) *Interactor {
	return &Interactor{
		storage: storage,
	}
}

func (i *Interactor) Call(ctx context.Context, id string) error {
	return i.storage.DeleteByID(ctx, id)
}
