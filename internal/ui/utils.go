package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"sismo_widget/internal/usecase"
)

func AbrirURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("plataforma no compatible")
	}
	return err
}

func ObtenerPathsConfig() (configPath, historyPath string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configDir := filepath.Join(home, ".sismo_widget")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", "", err
	}
	return filepath.Join(configDir, "config.json"), filepath.Join(configDir, "sismos.log"), nil
}

func MostrarVentanaConfig(
	fyneApp fyne.App,
	configManager *usecase.ConfigManager,
	monitor *usecase.EarthquakeMonitor,
	widgetWin fyne.Window,
	primerInicio bool,
) {
	win := fyneApp.NewWindow("Configuración - IGP Monitor")
	win.Resize(fyne.NewSize(480, 620))
	win.CenterOnScreen()
	win.SetCloseIntercept(func() {
		fyneApp.Quit()
	})

	config, _ := configManager.GetConfig()

	lblInstrucciones := widget.NewLabelWithStyle("⚙️ CONFIGURACIÓN", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	inputLat := widget.NewEntry()
	inputLon := widget.NewEntry()
	if primerInicio {
		inputLat.SetText("")
		inputLon.SetText("")
	} else {
		inputLat.SetText(strconv.FormatFloat(config.UserLat, 'f', 6, 64))
		inputLon.SetText(strconv.FormatFloat(config.UserLon, 'f', 6, 64))
	}

	inputAlertaMag := widget.NewEntry()
	inputAlertaMag.SetText(strconv.FormatFloat(config.AlertaMagnitudMin, 'f', 1, 64))

	inputAlertaDist := widget.NewEntry()
	inputAlertaDist.SetText(strconv.FormatFloat(config.AlertaDistanciaMax, 'f', 0, 64))

	inputFiltroMag := widget.NewEntry()
	inputFiltroMag.SetText(strconv.FormatFloat(config.FiltroMagnitudMin, 'f', 1, 64))

	inputFiltroDist := widget.NewEntry()
	inputFiltroDist.SetText(strconv.FormatFloat(config.FiltroDistanciaMax, 'f', 0, 64))

	chkAlertaGlobal := widget.NewCheck("Alertar siempre si Magnitud ≥ 6 (sin importar distancia)", func(v bool) {})
	chkAlertaGlobal.SetChecked(config.AlertaGlobalMagnitud6)

	selectAudio := widget.NewSelect(GetAudioDeviceNames(), func(s string) {})
	if config.AudioDevice != "" {
		selectAudio.SetSelected(config.AudioDevice)
	} else {
		selectAudio.SetSelected("Predeterminado")
	}

	lblEstado := widget.NewLabel("Ajusta tus preferencias.")

	btnAuto := widget.NewButtonWithIcon("Autodetectar por IP", theme.SearchIcon(), func() {
		detectarUbicacionPorIP(inputLat, inputLon, lblEstado, "🎯 ¡Ubicación actualizada!")
	})

	btnMapa := widget.NewButtonWithIcon("Seleccionar en Mapa", theme.SearchIcon(), func() {
		config, _ := configManager.GetConfig()
		if config.UserLat == 0 && config.UserLon == 0 {
			AbrirURL("https://www.openstreetmap.org/")
		} else {
			AbrirURL(fmt.Sprintf("https://www.openstreetmap.org/?mlat=%f&mlon=%f#map=15/%f/%f",
				config.UserLat, config.UserLon, config.UserLat, config.UserLon))
		}
	})

	btnProbarAlerta := widget.NewButtonWithIcon("Probar Alerta", theme.WarningIcon(), func() {
		lat, errLat := strconv.ParseFloat(inputLat.Text, 64)
		lon, errLon := strconv.ParseFloat(inputLon.Text, 64)
		if errLat != nil || errLon != nil {
			dialog.ShowError(fmt.Errorf("verifica las coordenadas: lat o lon no es válida"), win)
			return
		}

		alertaMag, err1 := strconv.ParseFloat(inputAlertaMag.Text, 64)
		alertaDist, err2 := strconv.ParseFloat(inputAlertaDist.Text, 64)
		filtroMag, err3 := strconv.ParseFloat(inputFiltroMag.Text, 64)
		filtroDist, err4 := strconv.ParseFloat(inputFiltroDist.Text, 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			dialog.ShowError(fmt.Errorf("verifica los valores numéricos"), win)
			return
		}

		nuevoConfig := configManager.GetConfigOrDefault()
		nuevoConfig.UserLat = lat
		nuevoConfig.UserLon = lon
		nuevoConfig.AlertaMagnitudMin = alertaMag
		nuevoConfig.AlertaDistanciaMax = alertaDist
		nuevoConfig.FiltroMagnitudMin = filtroMag
		nuevoConfig.FiltroDistanciaMax = filtroDist
		nuevoConfig.AlertaGlobalMagnitud6 = chkAlertaGlobal.Checked
		nuevoConfig.AudioDevice = selectAudio.Selected
		configManager.SaveConfig(nuevoConfig)
		monitor.UpdateConfig(nuevoConfig)

		win.Hide()
		if widgetWin != nil {
			mostrarVentanaWidget(widgetWin, configManager)
			sismoSimulado := monitor.CreateSimulatedEarthquake()
			sismoSimulado.CalcularDistancia(nuevoConfig.UserLat, nuevoConfig.UserLon)
			go func() {
				time.Sleep(500 * time.Millisecond)
				fyne.Do(func() {
					alertWin := NewFullscreenAlert(fyneApp, configManager)
					alertWin.Show(sismoSimulado)
				})
			}()
		} else {
			MostrarWidget(fyneApp, configManager, monitor, true)
		}
	})

	btnHistorial := widget.NewButtonWithIcon("Ver Historial", theme.ListIcon(), func() {
		MostrarVentanaHistorial(fyneApp, monitor)
	})

	btnIniciar := widget.NewButtonWithIcon("Iniciar Widget", theme.MediaPlayIcon(), func() {
		lat, errLat := strconv.ParseFloat(inputLat.Text, 64)
		lon, errLon := strconv.ParseFloat(inputLon.Text, 64)
		if errLat != nil || errLon != nil {
			dialog.ShowError(fmt.Errorf("verifica las coordenadas: lat o lon no es válida"), win)
			return
		}

		alertaMag, err1 := strconv.ParseFloat(inputAlertaMag.Text, 64)
		alertaDist, err2 := strconv.ParseFloat(inputAlertaDist.Text, 64)
		filtroMag, err3 := strconv.ParseFloat(inputFiltroMag.Text, 64)
		filtroDist, err4 := strconv.ParseFloat(inputFiltroDist.Text, 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			dialog.ShowError(fmt.Errorf("verifica los valores numéricos"), win)
			return
		}

		nuevoConfig := configManager.GetConfigOrDefault()
		nuevoConfig.UserLat = lat
		nuevoConfig.UserLon = lon
		nuevoConfig.AlertaMagnitudMin = alertaMag
		nuevoConfig.AlertaDistanciaMax = alertaDist
		nuevoConfig.FiltroMagnitudMin = filtroMag
		nuevoConfig.FiltroDistanciaMax = filtroDist
		nuevoConfig.AlertaGlobalMagnitud6 = chkAlertaGlobal.Checked
		nuevoConfig.AudioDevice = selectAudio.Selected
		configManager.SaveConfig(nuevoConfig)
		monitor.UpdateConfig(nuevoConfig)

		win.Hide()
		if widgetWin != nil {
			mostrarVentanaWidget(widgetWin, configManager)
		} else {
			MostrarWidget(fyneApp, configManager, monitor, false)
		}
	})

	contenido := container.NewVBox(
		lblInstrucciones,
		widget.NewSeparator(),
		widget.NewForm(
			widget.NewFormItem("📍 Latitud:", inputLat),
			widget.NewFormItem("📍 Longitud:", inputLon),
		),
		container.NewGridWithColumns(2, btnAuto, btnMapa),
		widget.NewSeparator(),
		widget.NewLabel("⚠️ Alertas (ejecutar alerta):"),
		widget.NewForm(
			widget.NewFormItem("Mín. Magnitud:", inputAlertaMag),
			widget.NewFormItem("Máx. Distancia (km):", inputAlertaDist),
		),
		widget.NewSeparator(),
		chkAlertaGlobal,
		widget.NewSeparator(),
		widget.NewLabel("🔊 Dispositivo de audio:"),
		selectAudio,
		widget.NewSeparator(),
		widget.NewLabel("📋 Filtros (registrar en historial):"),
		widget.NewForm(
			widget.NewFormItem("Mín. Magnitud:", inputFiltroMag),
			widget.NewFormItem("Máx. Distancia (km):", inputFiltroDist),
		),
		widget.NewSeparator(),
		lblEstado,
		widget.NewSeparator(),
		container.NewGridWithColumns(3, btnHistorial, btnProbarAlerta, btnIniciar),
	)

	win.SetContent(container.NewPadded(contenido))
	win.Show()
	if primerInicio {
		detectarUbicacionPorIP(inputLat, inputLon, lblEstado, "🎯 ¡Ubicación detectada!")
	}
}

func MostrarVentanaHistorial(fyneApp fyne.App, monitor *usecase.EarthquakeMonitor) {
	winHistorial := fyneApp.NewWindow("Historial de Sismos")
	winHistorial.Resize(fyne.NewSize(550, 500))
	winHistorial.CenterOnScreen()

	historial, _ := monitor.GetHistory()
	var items []string
	for _, sismo := range historial {
		distStr := ""
		if sismo.Distancia > 0 {
			distStr = fmt.Sprintf(" | %.0f km", sismo.Distancia)
		}
		fechaStr := sismo.FechaLocal
		if fechaStr == "" {
			fechaStr = sismo.FechaHora[:10]
		}
		horaStr := sismo.HoraLocal
		if horaStr == "" {
			horaStr = sismo.FechaHora[11:19]
		}
		items = append(items, fmt.Sprintf("M%.1f - %s\n%s %s%s",
			sismo.Magnitud, sismo.Referencia, fechaStr, horaStr, distStr))
	}

	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject {
			return widget.NewLabel("M4.5 - Lugar")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(items[i])
			o.(*widget.Label).Wrapping = fyne.TextWrapWord
		},
	)

	btnLimpiar := widget.NewButtonWithIcon("Limpiar Historial", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Confirmar", "¿Estás seguro de que quieres limpiar el historial?", func(confirm bool) {
			if confirm {
				monitor.ClearHistory()
				winHistorial.Close()
				MostrarVentanaHistorial(fyneApp, monitor)
			}
		}, winHistorial)
	})

	contenido := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("📜 HISTORIAL DE SISMOS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
		),
		container.NewPadded(btnLimpiar),
		nil,
		nil,
		list,
	)

	winHistorial.SetContent(contenido)
	winHistorial.Show()
}

var systrayIniciado bool = false

func guardarPosicionVentana(cm *usecase.ConfigManager) {
	x, y := GetWindowPosition()
	cfg, _ := cm.GetConfig()
	cfg.WindowX = x
	cfg.WindowY = y
	cm.SaveConfig(cfg)
}

func mostrarVentanaWidget(widgetWin fyne.Window, cm *usecase.ConfigManager) {
	cfg, _ := cm.GetConfig()
	if cfg.WindowX != 0 || cfg.WindowY != 0 {
		SetWindowPosition(cfg.WindowX, cfg.WindowY)
	}
	widgetWin.Show()
	if cfg.WindowX != 0 || cfg.WindowY != 0 {
		SetWindowPosition(cfg.WindowX, cfg.WindowY)
	}
}

func detectarUbicacionPorIP(inputLat, inputLon *widget.Entry, lblEstado *widget.Label, mensajeExito string) {
	lblEstado.SetText("Buscando ubicación...")
	go func() {
		resp, err := http.Get("https://ip-api.com/json/")
		if err == nil {
			var geo struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			}
			if json.NewDecoder(resp.Body).Decode(&geo) == nil && geo.Lat != 0 {
				fyne.Do(func() {
					inputLat.SetText(strconv.FormatFloat(geo.Lat, 'f', 6, 64))
					inputLon.SetText(strconv.FormatFloat(geo.Lon, 'f', 6, 64))
					lblEstado.SetText(mensajeExito)
				})
			} else {
				fyne.Do(func() { lblEstado.SetText("No se pudo detectar la IP.") })
			}
			resp.Body.Close()
		} else {
			fyne.Do(func() { lblEstado.SetText("Error de conexión.") })
		}
	}()
}

func iniciarSystray(fyneApp fyne.App, widgetWin fyne.Window, configManager *usecase.ConfigManager) {
	if systrayIniciado {
		return
	}
	systrayIniciado = true

	desk, ok := fyneApp.(desktop.App)
	if !ok {
		return
	}

	menu := fyne.NewMenu("IGP Sismo Monitor",
		fyne.NewMenuItem("Mostrar", func() {
			mostrarVentanaWidget(widgetWin, configManager)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Salir", func() {
			guardarPosicionVentana(configManager)
			go func() {
				time.Sleep(200 * time.Millisecond)
				os.Exit(0)
			}()
			fyneApp.Quit()
		}),
	)

	desk.SetSystemTrayMenu(menu)
	desk.SetSystemTrayIcon(fyneApp.Icon())
}
