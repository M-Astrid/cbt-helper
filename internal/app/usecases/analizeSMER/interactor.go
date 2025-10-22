package analizeSMER

import (
	"context"
	"errors"

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

func (i *Interactor) Call(ctx context.Context, id string, notes string) error {
	s, err := i.storage.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if s == nil {
		return errors.New("запись СМЭР не найдена")
	}
	s.Unstructured = entity.NewUnstructured(notes)
	return i.storage.Save(ctx, s)
}
