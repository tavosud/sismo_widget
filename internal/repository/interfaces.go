package repository

import "sismo_widget/internal/entity"

type ConfigRepository interface {
	Load() (entity.Config, error)
	Save(config entity.Config) error
}

type HistoryRepository interface {
	Load() ([]entity.Sismo, error)
	Save(sismos []entity.Sismo) error
	Add(sismo entity.Sismo) error
}

type IGPApiRepository interface {
	GetUltimoSismo() (*entity.Sismo, error)
}
