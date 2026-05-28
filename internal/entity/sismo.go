package entity

import (
	"math"
	"strconv"
	"time"
)

type Sismo struct {
	ID          int       `json:"id"`
	Codigo      string    `json:"codigo"`
	FechaHora   string    `json:"fecha_hora"`
	MagnitudStr string    `json:"magnitud"`
	Referencia  string    `json:"referencia"`
	LatitudStr  string    `json:"latitud"`
	LongitudStr string    `json:"longitud"`
	Profundidad float64   `json:"profundidad"`
	Distancia   float64   `json:"distancia,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Magnitud    float64   `json:"magnitud_num"`
	Latitud     float64   `json:"latitud_num"`
	Longitud    float64   `json:"longitud_num"`
	FechaLocal  string    `json:"fecha_local,omitempty"`
	HoraLocal   string    `json:"hora_local,omitempty"`
	Intensidad  string    `json:"intensidad,omitempty"`
}

func (s *Sismo) CalcularDistancia(userLat, userLon float64) {
	const R = 6371.0
	lat1 := userLat * math.Pi / 180.0
	lat2 := s.Latitud * math.Pi / 180.0
	dLat := (s.Latitud - userLat) * math.Pi / 180.0
	dLon := (s.Longitud - userLon) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	s.Distancia = R * (2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a)))
}

func (s *Sismo) EnsureParsed() {
	if s.Magnitud == 0 && s.MagnitudStr != "" {
		if mag, err := strconv.ParseFloat(s.MagnitudStr, 64); err == nil {
			s.Magnitud = mag
		}
	}
	if s.Latitud == 0 && s.LatitudStr != "" {
		if lat, err := strconv.ParseFloat(s.LatitudStr, 64); err == nil {
			s.Latitud = lat
		}
	}
	if s.Longitud == 0 && s.LongitudStr != "" {
		if lon, err := strconv.ParseFloat(s.LongitudStr, 64); err == nil {
			s.Longitud = lon
		}
	}
	if (s.FechaLocal == "" || s.HoraLocal == "") && s.FechaHora != "" {
		if t, err := time.Parse(time.RFC3339, s.FechaHora); err == nil {
			local := t.Local()
			s.FechaLocal = local.Format("02/01/2006")
			s.HoraLocal = local.Format("15:04:05")
		}
	}
}
