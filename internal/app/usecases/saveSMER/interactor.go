package saveSMERUsecase

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

func (i *Interactor) Call(ctx context.Context, smer *entity.SMEREntry) error {
	return i.storage.Save(ctx, smer)
}
