package port

import "github.com/M-Astrid/cbt-helper/internal/app/dto"

type AIAssistant interface {
	AnalizeSMER(smer string) (*dto.AnalizeSMERResponse, error)
}
