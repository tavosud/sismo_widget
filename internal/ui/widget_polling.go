package ui

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"

	"sismo_widget/internal/entity"
	"sismo_widget/internal/usecase"
)

func simularAlerta(fyneApp fyne.App, configManager *usecase.ConfigManager, monitor *usecase.EarthquakeMonitor) {
	sismoSimulado := monitor.CreateSimulatedEarthquake()
	cfg, err := configManager.GetConfig()
	if err != nil {
		log.Printf("error leyendo config para simulacion: %v", err)
	}
	sismoSimulado.CalcularDistancia(cfg.UserLat, cfg.UserLon)
	go func() {
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			alertWin := NewFullscreenAlert(fyneApp)
			alertWin.Show(sismoSimulado)
		})
	}()
}

func (w *widgetUI) handleSimulation() {
	sismoSimulado := w.monitor.CreateSimulatedEarthquake()
	cfg, err := w.configMgr.GetConfig()
	if err != nil {
		log.Printf("error leyendo config: %v", err)
	}
	sismoSimulado.CalcularDistancia(cfg.UserLat, cfg.UserLon)
	w.actualizarUI(sismoSimulado)
	go func() {
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			w.mostrarAlerta(sismoSimulado)
		})
	}()
}

func (w *widgetUI) startPolling() {
	go func() {
		for {
			sismo, esNuevo, err := w.monitor.CheckForNewEarthquake()
			if err == nil {
				fyne.Do(func() {
					w.actualizarUI(sismo)
				})
				if esNuevo && w.monitor.IsAlert(sismo) {
					fyne.Do(func() {
						w.mostrarAlerta(sismo)
					})
				}
			} else {
				fyne.Do(func() {
					w.txtUbicacion.Text = "Error: " + err.Error()
					w.txtUbicacion.Refresh()
				})
			}
			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		time.Sleep(300 * time.Millisecond)
		sismo, _, err := w.monitor.CheckForNewEarthquake()
		if err == nil {
			fyne.Do(func() {
				w.actualizarUI(sismo)
			})
		}
	}()
}

func (w *widgetUI) actualizarUI(sismo *entity.Sismo) {
	if sismo == nil {
		return
	}
	w.mapLat = sismo.Latitud
	w.mapLon = sismo.Longitud

	cfg, err := w.configMgr.GetConfig()
	if err != nil {
		log.Printf("error leyendo config: %v", err)
	}
	sismo.CalcularDistancia(cfg.UserLat, cfg.UserLon)

	w.txtMagnitud.Text = fmt.Sprintf("M %.1f", sismo.Magnitud)
	w.txtUbicacion.Text = sismo.Referencia

	var headerColor, headerTextColor color.Color
	mag := sismo.Magnitud
	switch {
	case mag >= 6:
		headerColor = w.clrs.rojo
		headerTextColor = color.White
	case mag >= 4.5:
		headerColor = w.clrs.amarillo
		headerTextColor = w.clrs.negro
	default:
		headerColor = w.clrs.verde
		headerTextColor = color.White
	}
	w.headerBg.FillColor = headerColor
	w.headerBg.Refresh()
	w.txtMagnitud.Color = headerTextColor
	w.txtUbicacion.Color = headerTextColor
	w.txtFechaHeader.Color = headerTextColor

	fechaStr := ""
	if sismo.FechaLocal != "" && sismo.HoraLocal != "" {
		fechaStr = fmt.Sprintf("%s a las %s hrs", sismo.FechaLocal, sismo.HoraLocal)
	} else {
		fechaStr = sismo.FechaHora
	}
	w.txtFechaHeader.Text = fechaStr

	w.txtMagnitud.Refresh()
	w.txtUbicacion.Refresh()
	w.txtFechaHeader.Refresh()

	w.txtProfundidad.Text = fmt.Sprintf("%.0f km", sismo.Profundidad)
	intensidad := sismo.Intensidad
	if intensidad == "" {
		intensidad = "--"
	}
	w.txtIntensidad.Text = intensidad
	w.txtDistancia.Text = fmt.Sprintf("%.0f km", sismo.Distancia)

	w.txtProfundidad.Refresh()
	w.txtIntensidad.Refresh()
	w.txtDistancia.Refresh()

	w.cargarMapa(sismo.Latitud, sismo.Longitud)
}

func (w *widgetUI) mostrarAlerta(sismo *entity.Sismo) {
	alertWin := NewFullscreenAlert(w.fyneApp)
	alertWin.Show(sismo)
	w.fyneApp.SendNotification(&fyne.Notification{
		Title:   "¡ALERTA SÍSMICA!",
		Content: fmt.Sprintf("M%.1f - %s (%.0f km)", sismo.Magnitud, sismo.Referencia, sismo.Distancia),
	})
}
