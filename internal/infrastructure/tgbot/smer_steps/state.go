package smer_steps

import (
	"time"

	"github.com/M-Astrid/cbt-helper/internal/domain/entity"
)

type UserState struct {
	Status             int
	SMER               *entity.SMEREntry
	SMERSteps          []StepI
	CurrentSMERStepIdx int
	StartDate          time.Time
	EndDate            time.Time
}
