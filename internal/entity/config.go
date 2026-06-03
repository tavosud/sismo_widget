package entity

type Config struct {
	UserLat              float64 `json:"user_lat"`
	UserLon              float64 `json:"user_lon"`
	AlertaMagnitudMin    float64 `json:"alerta_magnitud_min"`
	AlertaDistanciaMax   float64 `json:"alerta_distancia_max"`
	FiltroMagnitudMin    float64 `json:"filtro_magnitud_min"`
	FiltroDistanciaMax   float64 `json:"filtro_distancia_max"`
	WindowX              int     `json:"window_x,omitempty"`
	WindowY              int     `json:"window_y,omitempty"`
	AlertaGlobalMagnitud6 bool   `json:"alerta_global_magnitud_6"`
	AudioDevice          string  `json:"audio_device,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		UserLat:               -12.112,
		UserLon:               -77.014,
		AlertaMagnitudMin:     4.5,
		AlertaDistanciaMax:    200,
		FiltroMagnitudMin:     2.0,
		FiltroDistanciaMax:    1000,
		AlertaGlobalMagnitud6: true,
	}
}
