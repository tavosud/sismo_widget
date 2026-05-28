package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sismo_widget/internal/entity"
	"strconv"
	"time"
)

type igpApiRepository struct {
	client *http.Client
}

func NewIGPApiRepository() *igpApiRepository {
	return &igpApiRepository{
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

type apiSismo struct {
	ID          int     `json:"id"`
	Codigo      string  `json:"codigo"`
	FechaHora   string  `json:"fecha_hora"`
	MagnitudStr string  `json:"magnitud"`
	Referencia  string  `json:"referencia"`
	LatitudStr  string  `json:"latitud"`
	LongitudStr string  `json:"longitud"`
	Profundidad  float64 `json:"profundidad"`
	Intensidades string  `json:"intensidades"`
}

func (r *igpApiRepository) GetUltimoSismo() (*entity.Sismo, error) {
	req, err := http.NewRequest("GET", "https://ultimosismo.igp.gob.pe/api/ultimo-sismo", nil)
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de conexión: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("servidor respondió con código: %d", resp.StatusCode)
	}

	var apiS apiSismo
	err = json.NewDecoder(resp.Body).Decode(&apiS)
	if err != nil {
		return nil, fmt.Errorf("error decodificando datos: %v", err)
	}

	sismo := &entity.Sismo{
		ID:          apiS.ID,
		Codigo:      apiS.Codigo,
		FechaHora:   apiS.FechaHora,
		MagnitudStr: apiS.MagnitudStr,
		Referencia:  apiS.Referencia,
		LatitudStr:  apiS.LatitudStr,
		LongitudStr: apiS.LongitudStr,
		Profundidad: apiS.Profundidad,
		Intensidad:  apiS.Intensidades,
	}

	if err := parseSismoFields(sismo); err != nil {
		return nil, fmt.Errorf("error parseando campos: %v", err)
	}

	return sismo, nil
}

func parseSismoFields(s *entity.Sismo) error {
	var err error
	if s.Magnitud, err = strconv.ParseFloat(s.MagnitudStr, 64); err != nil {
		return fmt.Errorf("magnitud inválida: %v", err)
	}
	if s.Latitud, err = strconv.ParseFloat(s.LatitudStr, 64); err != nil {
		return fmt.Errorf("latitud inválida: %v", err)
	}
	if s.Longitud, err = strconv.ParseFloat(s.LongitudStr, 64); err != nil {
		return fmt.Errorf("longitud inválida: %v", err)
	}

	t, err := time.Parse(time.RFC3339, s.FechaHora)
	if err == nil {
		local := t.Local()
		s.FechaLocal = local.Format("02/01/2006")
		s.HoraLocal = local.Format("15:04:05")
	} else {
		s.FechaLocal = ""
		s.HoraLocal = s.FechaHora
	}

	return nil
}
