package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sismo_widget/internal/entity"
)

type configFileRepository struct {
	filePath string
}

func NewConfigFileRepository(filePath string) *configFileRepository {
	return &configFileRepository{filePath: filePath}
}

func (r *configFileRepository) Load() (entity.Config, error) {
	config := entity.DefaultConfig()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return config, nil
	}

	err = json.Unmarshal(data, &config)
	return config, err
}

func (r *configFileRepository) Save(config entity.Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
}
