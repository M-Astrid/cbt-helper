package port

import (
	"context"
	"time"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
)

type SMERStorageI interface {
	Save(ctx context.Context, entry *entity.SMEREntry) error
	GetByUserID(ctx context.Context, id int64, startDate time.Time, endDate time.Time) ([]*entity.SMEREntry, error)
	GetByID(ctx context.Context, id string) (*entity.SMEREntry, error)
	DeleteByID(ctx context.Context, id string) error
}
