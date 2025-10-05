package steps

type UserState struct {
	Status              int
	SMERID              int64
	SMERSteps           []StepI
	CurrentSMERStepIdx  int
	CurrentSMERStepType int
}
