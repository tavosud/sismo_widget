package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sismo_widget/internal/entity"
)

type mockConfigRepo struct{ mock.Mock }

func (m *mockConfigRepo) Load() (entity.Config, error) {
	args := m.Called()
	return args.Get(0).(entity.Config), args.Error(1)
}

func (m *mockConfigRepo) Save(config entity.Config) error {
	args := m.Called(config)
	return args.Error(0)
}

func TestGetConfigOrDefault_LoadFails(t *testing.T) {
	repo := new(mockConfigRepo)
	repo.On("Load").Return(entity.Config{}, errors.New("not found"))

	cm := NewConfigManager(repo)
	cfg := cm.GetConfigOrDefault()

	assert.Equal(t, -12.112, cfg.UserLat)
	assert.Equal(t, -77.014, cfg.UserLon)
	assert.True(t, cfg.AlertaGlobalMagnitud6)
}

func TestGetConfigOrDefault_LoadSucceeds(t *testing.T) {
	expected := entity.Config{
		UserLat:  -10.0,
		UserLon:  -75.0,
		AlertaGlobalMagnitud6: false,
	}
	repo := new(mockConfigRepo)
	repo.On("Load").Return(expected, nil)

	cm := NewConfigManager(repo)
	cfg := cm.GetConfigOrDefault()

	assert.Equal(t, -10.0, cfg.UserLat)
	assert.Equal(t, -75.0, cfg.UserLon)
	assert.False(t, cfg.AlertaGlobalMagnitud6)
}

func TestGetConfig_Success(t *testing.T) {
	expected := entity.Config{UserLat: -12.112, UserLon: -77.014}
	repo := new(mockConfigRepo)
	repo.On("Load").Return(expected, nil)

	cm := NewConfigManager(repo)
	cfg, err := cm.GetConfig()

	assert.NoError(t, err)
	assert.Equal(t, expected, cfg)
}

func TestGetConfig_Error(t *testing.T) {
	repo := new(mockConfigRepo)
	repo.On("Load").Return(entity.Config{}, errors.New("read error"))

	cm := NewConfigManager(repo)
	_, err := cm.GetConfig()

	assert.Error(t, err)
}

func TestSaveConfig(t *testing.T) {
	cfg := entity.Config{UserLat: -12.0, UserLon: -77.0}
	repo := new(mockConfigRepo)
	repo.On("Save", cfg).Return(nil)

	cm := NewConfigManager(repo)
	err := cm.SaveConfig(cfg)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSaveConfig_Error(t *testing.T) {
	cfg := entity.Config{UserLat: -12.0}
	repo := new(mockConfigRepo)
	repo.On("Save", cfg).Return(errors.New("write error"))

	cm := NewConfigManager(repo)
	err := cm.SaveConfig(cfg)

	assert.Error(t, err)
}

func TestNewConfigManager(t *testing.T) {
	repo := new(mockConfigRepo)
	cm := NewConfigManager(repo)
	assert.NotNil(t, cm)
	assert.Equal(t, repo, cm.configRepo)
}
