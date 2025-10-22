package providers

import (
	"github.com/M-Astrid/cbt-helper/internal/app/port"
	deleteSMERUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/del_smer"
	getSingleSMERUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/get_smer"
	getUserSMERsUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/get_smers"
	saveSMERUsecase "github.com/M-Astrid/cbt-helper/internal/app/usecases/save_smer"
)

func ProvideGetSingleSMERUsecase(storage port.SMERStorage) *getSingleSMERUsecase.Interactor {
	return getSingleSMERUsecase.NewInteractor(storage)
}

func ProvideGetUserSMERsUsecase(storage port.SMERStorage) *getUserSMERsUsecase.Interactor {
	return getUserSMERsUsecase.NewInteractor(storage)
}

func ProvideSaveSMERUsecase(storage port.SMERStorage) *saveSMERUsecase.Interactor {
	return saveSMERUsecase.NewInteractor(storage)
}

func ProvideDeleteSMERUsecase(storage port.SMERStorage) *deleteSMERUsecase.Interactor {
	return deleteSMERUsecase.NewInteractor(storage)
}
