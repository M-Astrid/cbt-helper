package port

import (
	"context"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
)

type SMERStorageI interface {
	Save(ctx context.Context, entry *entity.SMEREntry) error
	GetByUserID(ctx context.Context, id int64) ([]*entity.SMEREntry, error)
}
