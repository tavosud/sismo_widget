package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sismo_widget/internal/entity"
)

type mockIGPRepo struct{ mock.Mock }

func (m *mockIGPRepo) GetUltimoSismo() (*entity.Sismo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Sismo), args.Error(1)
}

type mockHistoryRepo struct{ mock.Mock }

func (m *mockHistoryRepo) Load() ([]entity.Sismo, error) {
	args := m.Called()
	return args.Get(0).([]entity.Sismo), args.Error(1)
}

func (m *mockHistoryRepo) Save(sismos []entity.Sismo) error {
	args := m.Called(sismos)
	return args.Error(0)
}

func (m *mockHistoryRepo) Add(sismo entity.Sismo) error {
	args := m.Called(sismo)
	return args.Error(0)
}

func TestIsAlert_GlobalMagnitud6_True(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{AlertaGlobalMagnitud6: true})
	s := &entity.Sismo{Magnitud: 6.0, Distancia: 1000}
	assert.True(t, em.IsAlert(s))
}

func TestIsAlert_GlobalMagnitud6_FalseBelowThreshold(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{
		AlertaGlobalMagnitud6: false,
		AlertaMagnitudMin:     4.5,
		AlertaDistanciaMax:    200,
	})
	s := &entity.Sismo{Magnitud: 4.0, Distancia: 100}
	assert.False(t, em.IsAlert(s))
}

func TestIsAlert_WithinThresholds(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{
		AlertaGlobalMagnitud6: false,
		AlertaMagnitudMin:     4.5,
		AlertaDistanciaMax:    200,
	})
	s := &entity.Sismo{Magnitud: 5.0, Distancia: 150}
	assert.True(t, em.IsAlert(s))
}

func TestIsAlert_ExceedsDistance(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{
		AlertaGlobalMagnitud6: false,
		AlertaMagnitudMin:     4.5,
		AlertaDistanciaMax:    200,
	})
	s := &entity.Sismo{Magnitud: 5.0, Distancia: 300}
	assert.False(t, em.IsAlert(s))
}

func TestIsAlert_BelowMagnitude(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{
		AlertaGlobalMagnitud6: false,
		AlertaMagnitudMin:     4.5,
		AlertaDistanciaMax:    200,
	})
	s := &entity.Sismo{Magnitud: 3.5, Distancia: 50}
	assert.False(t, em.IsAlert(s))
}

func TestIsAlert_GlobalMagnitud6_WithHighMagnitudeAndDistance(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{
		AlertaGlobalMagnitud6: true,
		AlertaMagnitudMin:     4.5,
		AlertaDistanciaMax:    200,
	})
	s := &entity.Sismo{Magnitud: 7.0, Distancia: 9999}
	assert.True(t, em.IsAlert(s))
}

func TestCheckForNewEarthquake_APIFailure(t *testing.T) {
	igp := new(mockIGPRepo)
	igp.On("GetUltimoSismo").Return(nil, errors.New("network error"))

	em := NewEarthquakeMonitor(igp, nil, entity.Config{})
	_, _, err := em.CheckForNewEarthquake()
	assert.Error(t, err)
	igp.AssertExpectations(t)
}

func TestCheckForNewEarthquake_SameID(t *testing.T) {
	sismo := &entity.Sismo{ID: 1, Magnitud: 4.5}
	igp := new(mockIGPRepo)
	igp.On("GetUltimoSismo").Return(sismo, nil)

	// High filter to avoid historyRepo nil dereference
	em := NewEarthquakeMonitor(igp, nil, entity.Config{
		UserLat: 0, UserLon: 0,
		FiltroMagnitudMin: 10,
	})
	// First call sets ultimoID
	_, firstNew, err := em.CheckForNewEarthquake()
	assert.NoError(t, err)
	assert.True(t, firstNew)

	// Second call with same ID should return false
	result, secondNew, err := em.CheckForNewEarthquake()
	assert.NoError(t, err)
	assert.False(t, secondNew)
	assert.Equal(t, sismo, result)
	igp.AssertNumberOfCalls(t, "GetUltimoSismo", 2)
}

func TestCheckForNewEarthquake_NewIDPassesFilter(t *testing.T) {
	sismo := &entity.Sismo{
		ID: 42, Magnitud: 5.0,
		Latitud: -12.0, Longitud: -77.0,
	}
	igp := new(mockIGPRepo)
	igp.On("GetUltimoSismo").Return(sismo, nil)

	history := new(mockHistoryRepo)
	history.On("Add", *sismo).Return(nil)

	em := NewEarthquakeMonitor(igp, history, entity.Config{
		UserLat:            -12.0,
		UserLon:            -77.0,
		FiltroMagnitudMin:  2.0,
		FiltroDistanciaMax: 1000,
	})

	result, isNew, err := em.CheckForNewEarthquake()
	assert.NoError(t, err)
	assert.True(t, isNew)
	assert.Equal(t, sismo, result)
	history.AssertExpectations(t)
}

func TestCheckForNewEarthquake_NewIDFailsFilter(t *testing.T) {
	sismo := &entity.Sismo{
		ID: 42, Magnitud: 0.5,
		Latitud: -12.0, Longitud: -77.0,
	}
	igp := new(mockIGPRepo)
	igp.On("GetUltimoSismo").Return(sismo, nil)

	em := NewEarthquakeMonitor(igp, nil, entity.Config{
		UserLat:            -12.0,
		UserLon:            -77.0,
		FiltroMagnitudMin:  2.0,
		FiltroDistanciaMax: 1000,
	})

	result, isNew, err := em.CheckForNewEarthquake()
	assert.NoError(t, err)
	assert.True(t, isNew)
	assert.Equal(t, sismo, result)
}

func TestCreateSimulatedEarthquake_ReturnsExpectedValues(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{UserLat: -12.112, UserLon: -77.014})
	s := em.CreateSimulatedEarthquake()

	assert.Equal(t, 99999, s.ID)
	assert.Equal(t, "SIM-001", s.Codigo)
	assert.Equal(t, 5.8, s.Magnitud)
	assert.Equal(t, 10.0, s.Distancia)
	assert.Equal(t, -12.012, s.Latitud)  // -12.112 + 0.1
	assert.Equal(t, -76.914, s.Longitud) // -77.014 + 0.1
	assert.Contains(t, s.Referencia, "SIMULACRO")
}

func TestUpdateConfig(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{UserLat: 0, UserLon: 0})
	em.UpdateConfig(entity.Config{UserLat: -12.112, UserLon: -77.014})

	assert.Equal(t, -12.112, em.config.UserLat)
	assert.Equal(t, -77.014, em.config.UserLon)
}

func TestIsAlert_GlobalFlagOff_EdgeValues(t *testing.T) {
	em := NewEarthquakeMonitor(nil, nil, entity.Config{
		AlertaGlobalMagnitud6: false,
		AlertaMagnitudMin:     4.5,
		AlertaDistanciaMax:    200,
	})
	tests := []struct {
		mag    float64
		dist   float64
		expect bool
	}{
		{4.5, 200, true},   // exactly at thresholds
		{4.5, 201, false},  // exactly at mag, over distance
		{4.4, 200, false},  // under mag, at distance
		{6.0, 999, false},  // high mag but over distance (global flag off)
	}
	for _, tc := range tests {
		s := &entity.Sismo{Magnitud: tc.mag, Distancia: tc.dist}
		assert.Equal(t, tc.expect, em.IsAlert(s), "mag=%.1f dist=%.0f", tc.mag, tc.dist)
	}
}

func TestCheckForNewEarthquake_CalculatesDistance(t *testing.T) {
	sismo := &entity.Sismo{
		ID: 100, Magnitud: 4.0,
		Latitud: -12.5, Longitud: -77.5,
	}
	igp := new(mockIGPRepo)
	igp.On("GetUltimoSismo").Return(sismo, nil)

	history := new(mockHistoryRepo)
	history.On("Add", mock.Anything).Return(nil)

	em := NewEarthquakeMonitor(igp, history, entity.Config{
		UserLat: -12.112, UserLon: -77.014,
		FiltroMagnitudMin: 0, FiltroDistanciaMax: 10000,
	})

	result, _, err := em.CheckForNewEarthquake()
	assert.NoError(t, err)
	assert.True(t, result.Distancia > 0)
	assert.True(t, result.Distancia < 100)
}

func TestGetHistory(t *testing.T) {
	history := new(mockHistoryRepo)
	expected := []entity.Sismo{{ID: 1}, {ID: 2}}
	history.On("Load").Return(expected, nil)

	em := NewEarthquakeMonitor(nil, history, entity.Config{})
	got, err := em.GetHistory()
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestClearHistory(t *testing.T) {
	history := new(mockHistoryRepo)
	history.On("Save", []entity.Sismo{}).Return(nil)

	em := NewEarthquakeMonitor(nil, history, entity.Config{})
	err := em.ClearHistory()
	assert.NoError(t, err)
}
