package dto

import "github.com/M-Astrid/cbt-helper/internal/domain/entity"

type AnalizeSMERRequest struct {
	smer *entity.SMEREntry `json:"smer"`
}

type AnalizeSMERResponse struct {
	Message string `json:"message"`
}
