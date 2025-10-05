package storage

import (
	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
)

type TriggerRepository interface {
	Save(trigger *entity.Trigger) error
	GetByID(id int64) (*entity.Trigger, error)
	GetBySMERID(id int64) (*entity.Trigger, error)
}
