package usecase

import (
	"sismo_widget/internal/entity"
	"sismo_widget/internal/repository"
)

type ConfigManager struct {
	configRepo repository.ConfigRepository
}

func NewConfigManager(configRepo repository.ConfigRepository) *ConfigManager {
	return &ConfigManager{configRepo: configRepo}
}

func (cm *ConfigManager) GetConfig() (entity.Config, error) {
	return cm.configRepo.Load()
}

func (cm *ConfigManager) GetConfigOrDefault() entity.Config {
	config, err := cm.configRepo.Load()
	if err != nil {
		return entity.DefaultConfig()
	}
	return config
}

func (cm *ConfigManager) SaveConfig(config entity.Config) error {
	return cm.configRepo.Save(config)
}
