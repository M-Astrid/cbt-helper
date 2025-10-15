package steps

import "github.com/M-Astrid/cbt-helper/internal/domain/entity"

type UserState struct {
	Status              int
	SMER                *entity.SMEREntry
	SMERSteps           []StepI
	CurrentSMERStepIdx  int
	CurrentSMERStepType int
}
