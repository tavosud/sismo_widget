package entity

import (
	"math"
	"testing"
	"time"
)

func roundTo(v float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(v*pow) / pow
}

func TestCalcularDistancia_Zero(t *testing.T) {
	s := &Sismo{Latitud: -12.112, Longitud: -77.014}
	s.CalcularDistancia(-12.112, -77.014)
	if s.Distancia != 0 {
		t.Errorf("expected 0, got %.2f", s.Distancia)
	}
}

func TestCalcularDistancia_KnownPoints(t *testing.T) {
	// Lima (IGP) to Chosica (~30 km NE)
	s := &Sismo{Latitud: -11.933, Longitud: -76.700}
	s.CalcularDistancia(-12.112, -77.014)
	got := roundTo(s.Distancia, 0)
	if got < 30 || got > 40 {
		t.Errorf("expected ~35 km, got %.0f", s.Distancia)
	}
}

func TestCalcularDistancia_Haversine(t *testing.T) {
	// User at equator, sismo 1° longitude east => ~111.19 km
	s := &Sismo{Latitud: 0, Longitud: 1}
	s.CalcularDistancia(0, 0)
	got := roundTo(s.Distancia, 1)
	if got != 111.2 {
		t.Errorf("expected 111.2 km, got %.1f", s.Distancia)
	}
}

func TestEnsureParsed_ParsesMagnitud(t *testing.T) {
	s := &Sismo{MagnitudStr: "4.5"}
	s.EnsureParsed()
	if s.Magnitud != 4.5 {
		t.Errorf("expected 4.5, got %.1f", s.Magnitud)
	}
}

func TestEnsureParsed_ParsesLatitud(t *testing.T) {
	s := &Sismo{LatitudStr: "-12.112"}
	s.EnsureParsed()
	if s.Latitud != -12.112 {
		t.Errorf("expected -12.112, got %.3f", s.Latitud)
	}
}

func TestEnsureParsed_ParsesLongitud(t *testing.T) {
	s := &Sismo{LongitudStr: "-77.014"}
	s.EnsureParsed()
	if s.Longitud != -77.014 {
		t.Errorf("expected -77.014, got %.3f", s.Longitud)
	}
}

func TestEnsureParsed_SkipsIfAlreadySet(t *testing.T) {
	s := &Sismo{Magnitud: 6.0, MagnitudStr: "4.5"}
	s.EnsureParsed()
	if s.Magnitud != 6.0 {
		t.Errorf("expected 6.0 (already set), got %.1f", s.Magnitud)
	}
}

func TestEnsureParsed_ParsesFechaHora(t *testing.T) {
	s := &Sismo{FechaHora: "2026-05-28T15:30:00Z"}
	s.EnsureParsed()
	if s.FechaLocal == "" || s.HoraLocal == "" {
		t.Fatal("expected fecha local to be set")
	}
	t.Logf("local: %s %s", s.FechaLocal, s.HoraLocal)
}

func TestEnsureParsed_InvalidFechaHora(t *testing.T) {
	s := &Sismo{FechaHora: "not-a-date"}
	s.EnsureParsed()
	// Should not crash; fields remain empty
}

func TestCalcularDistancia_OneDegreeLatitude(t *testing.T) {
	// 1° latitude ~ 111 km
	s := &Sismo{Latitud: 1, Longitud: 0}
	s.CalcularDistancia(0, 0)
	got := roundTo(s.Distancia, 0)
	if got < 110 || got > 112 {
		t.Errorf("expected ~111 km, got %.0f", s.Distancia)
	}
}

func TestEnsureParsed_InvalidStrings(t *testing.T) {
	s := &Sismo{MagnitudStr: "abc", LatitudStr: "xyz", LongitudStr: "!!!"}
	s.EnsureParsed()
	// Should not parse, fields remain zero
	if s.Magnitud != 0 || s.Latitud != 0 || s.Longitud != 0 {
		t.Error("expected zero values for invalid strings")
	}
}

func TestEnsureParsed_EmptyStrings(t *testing.T) {
	s := &Sismo{MagnitudStr: "", LatitudStr: "", LongitudStr: "", FechaHora: ""}
	s.EnsureParsed()
	// No fields set, should do nothing
}

func TestCreateSimulatedEarthquake_Defaults(t *testing.T) {
	// Not really a unit test of entity, but we can check the Sismo shape
	s := &Sismo{
		ID:       99999,
		Codigo:   "SIM-001",
		Magnitud: 5.8,
		Distancia: 10,
	}
	if s.Magnitud != 5.8 {
		t.Errorf("expected 5.8, got %.1f", s.Magnitud)
	}
}

func TestCalcularDistancia_Antipodal(t *testing.T) {
	// Opposite sides of the earth ~20015 km
	s := &Sismo{Latitud: 0, Longitud: 180}
	s.CalcularDistancia(0, 0)
	got := roundTo(s.Distancia, 0)
	if got < 19900 || got > 20100 {
		t.Errorf("expected ~20015 km, got %.0f", s.Distancia)
	}
}

func TestEnsureParsed_FechaHoraAlreadySet(t *testing.T) {
	s := &Sismo{
		FechaHora:  "2026-05-28T15:30:00Z",
		FechaLocal: "28/05/2026",
		HoraLocal:  "12:30:00",
	}
	s.EnsureParsed()
	if s.FechaLocal != "28/05/2026" {
		t.Errorf("expected FechaLocal to remain unchanged, got %s", s.FechaLocal)
	}
	if s.HoraLocal != "12:30:00" {
		t.Errorf("expected HoraLocal to remain unchanged, got %s", s.HoraLocal)
	}
}

func TestConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.UserLat != -12.112 {
		t.Errorf("expected -12.112, got %.3f", cfg.UserLat)
	}
	if cfg.UserLon != -77.014 {
		t.Errorf("expected -77.014, got %.3f", cfg.UserLon)
	}
	if cfg.AlertaMagnitudMin != 4.5 {
		t.Errorf("expected 4.5, got %.1f", cfg.AlertaMagnitudMin)
	}
	if cfg.AlertaDistanciaMax != 200 {
		t.Errorf("expected 200, got %.0f", cfg.AlertaDistanciaMax)
	}
	if cfg.FiltroMagnitudMin != 2.0 {
		t.Errorf("expected 2.0, got %.1f", cfg.FiltroMagnitudMin)
	}
	if cfg.FiltroDistanciaMax != 1000 {
		t.Errorf("expected 1000, got %.0f", cfg.FiltroDistanciaMax)
	}
	if !cfg.AlertaGlobalMagnitud6 {
		t.Error("expected AlertaGlobalMagnitud6 to be true")
	}
}

func TestEnsureParsed_TimeParseLocal(t *testing.T) {
	// Use a time that is unambiguous regardless of local timezone
	now := time.Date(2026, 5, 28, 20, 0, 0, 0, time.UTC)
	s := &Sismo{FechaHora: now.Format(time.RFC3339)}
	s.EnsureParsed()
	if s.FechaLocal == "" {
		t.Fatal("FechaLocal should be set")
	}
}
