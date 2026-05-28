package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sismo_widget/internal/entity"
	"sort"
	"time"
)

const maxHistorial = 50

type historyFileRepository struct {
	filePath string
}

func NewHistoryFileRepository(filePath string) *historyFileRepository {
	return &historyFileRepository{filePath: filePath}
}

func (r *historyFileRepository) Load() ([]entity.Sismo, error) {
	var sismos []entity.Sismo

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return sismos, nil
	}

	err = json.Unmarshal(data, &sismos)
	if err != nil {
		return sismos, err
	}

	for i := range sismos {
		sismos[i].EnsureParsed()
	}

	return sismos, nil
}

func (r *historyFileRepository) Save(sismos []entity.Sismo) error {
	if len(sismos) > maxHistorial {
		sismos = sismos[len(sismos)-maxHistorial:]
	}

	data, err := json.MarshalIndent(sismos, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
}

func (r *historyFileRepository) Add(sismo entity.Sismo) error {
	sismos, err := r.Load()
	if err != nil {
		return err
	}

	for _, s := range sismos {
		if s.ID == sismo.ID {
			return nil
		}
	}

	sismo.Timestamp = time.Now()
	sismos = append(sismos, sismo)

	sort.Slice(sismos, func(i, j int) bool {
		return sismos[i].Timestamp.After(sismos[j].Timestamp)
	})

	return r.Save(sismos)
}
