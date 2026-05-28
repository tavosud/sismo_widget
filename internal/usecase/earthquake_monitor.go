package usecase

import (
	"sismo_widget/internal/entity"
	"sismo_widget/internal/repository"
	"time"
)

type EarthquakeMonitor struct {
	igpRepo     repository.IGPApiRepository
	historyRepo repository.HistoryRepository
	config      entity.Config
	ultimoID    int
}

func NewEarthquakeMonitor(
	igpRepo repository.IGPApiRepository,
	historyRepo repository.HistoryRepository,
	config entity.Config,
) *EarthquakeMonitor {
	return &EarthquakeMonitor{
		igpRepo:     igpRepo,
		historyRepo: historyRepo,
		config:      config,
	}
}

func (em *EarthquakeMonitor) UpdateConfig(config entity.Config) {
	em.config = config
}

func (em *EarthquakeMonitor) CheckForNewEarthquake() (*entity.Sismo, bool, error) {
	sismo, err := em.igpRepo.GetUltimoSismo()
	if err != nil {
		return nil, false, err
	}

	if sismo.ID == em.ultimoID {
		return sismo, false, nil
	}

	em.ultimoID = sismo.ID
	sismo.CalcularDistancia(em.config.UserLat, em.config.UserLon)

	pasaFiltro := sismo.Magnitud >= em.config.FiltroMagnitudMin && sismo.Distancia <= em.config.FiltroDistanciaMax
	if pasaFiltro {
		em.historyRepo.Add(*sismo)
	}

	return sismo, true, nil
}

func (em *EarthquakeMonitor) IsAlert(sismo *entity.Sismo) bool {
	if em.config.AlertaGlobalMagnitud6 && sismo.Magnitud >= 6 {
		return true
	}
	return sismo.Magnitud >= em.config.AlertaMagnitudMin && sismo.Distancia <= em.config.AlertaDistanciaMax
}

func (em *EarthquakeMonitor) GetHistory() ([]entity.Sismo, error) {
	return em.historyRepo.Load()
}

func (em *EarthquakeMonitor) ClearHistory() error {
	return em.historyRepo.Save([]entity.Sismo{})
}

func (em *EarthquakeMonitor) CreateSimulatedEarthquake() *entity.Sismo {
	sismo := &entity.Sismo{
		ID:          99999,
		Codigo:      "SIM-001",
		FechaHora:   time.Now().Format(time.RFC3339),
		MagnitudStr: "5.8",
		Referencia:  "10 km al NE de tu ubicación - SIMULACRO",
		LatitudStr:  "0.0",
		LongitudStr: "0.0",
		Profundidad: 50,
		Magnitud:    5.8,
		Latitud:     em.config.UserLat + 0.1,
		Longitud:    em.config.UserLon + 0.1,
		FechaLocal:  time.Now().Format("02/01/2006"),
		HoraLocal:   time.Now().Format("15:04:05"),
		Distancia:   10,
	}
	return sismo
}
