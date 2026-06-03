package ui

import (
	"fmt"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"image/color"

	"sismo_widget/internal/entity"
	"sismo_widget/internal/usecase"
)

type FullscreenAlert struct {
	app         fyne.App
	win         fyne.Window
	fondo       *canvas.Rectangle
	running     bool
	cancelAudio chan struct{}
	configMgr   *usecase.ConfigManager
}

func NewFullscreenAlert(app fyne.App, configMgr *usecase.ConfigManager) *FullscreenAlert {
	alert := &FullscreenAlert{
		app:       app,
		configMgr: configMgr,
	}
	alert.createWindow()
	return alert
}

func (fa *FullscreenAlert) createWindow() {
	fa.win = fa.app.NewWindow("ALERTA SÍSMICA")
	fa.win.Hide()
}

func (fa *FullscreenAlert) Show(sismo *entity.Sismo) {
	fa.running = true
	fa.cancelAudio = make(chan struct{})

	var sitio, distStr, magStr string

	if sismo != nil {
		sitio = sismo.Referencia
		distStr = fmt.Sprintf("📏 Distancia: %.0f km", sismo.Distancia)
		magStr = fmt.Sprintf("🌊 Magnitud: M%.1f", sismo.Magnitud)
	} else {
		sitio = "10 km al NE de tu ubicación - SIMULACRO"
		distStr = "📏 Distancia: 10 km"
		magStr = "🌊 Magnitud: M5.8"
	}

	txtTitulo := canvas.NewText("¡ALERTA! ... Sismo en proceso", color.RGBA{R: 255, G: 50, B: 50, A: 255})
	txtTitulo.TextSize = 48
	txtTitulo.TextStyle = fyne.TextStyle{Bold: true}
	txtTitulo.Alignment = fyne.TextAlignCenter

	txtSitio := canvas.NewText(sitio, color.RGBA{R: 255, G: 200, B: 200, A: 255})
	txtSitio.TextSize = 28
	txtSitio.TextStyle = fyne.TextStyle{Bold: true}
	txtSitio.Alignment = fyne.TextAlignCenter

	txtDistancia := canvas.NewText(distStr, color.White)
	txtDistancia.TextSize = 20
	txtDistancia.Alignment = fyne.TextAlignCenter

	txtMagnitud := canvas.NewText(magStr, color.White)
	txtMagnitud.TextSize = 20
	txtMagnitud.Alignment = fyne.TextAlignCenter

	btnCerrar := widget.NewButton("🖱️ CERRAR ALERTA", func() {
		fa.cerrar()
	})
	btnCerrar.Importance = widget.HighImportance

	fa.fondo = canvas.NewRectangle(color.RGBA{R: 40, G: 10, B: 10, A: 255})

	contenido := container.NewVBox(
		txtTitulo,
		widget.NewSeparator(),
		txtSitio,
		container.NewGridWithColumns(2, txtDistancia, txtMagnitud),
		widget.NewSeparator(),
		btnCerrar,
	)

	content := container.NewMax(
		fa.fondo,
		container.NewCenter(container.NewPadded(contenido)),
	)

	fa.win.SetContent(content)
	fa.win.Resize(fyne.NewSize(1920, 1080))
	fa.win.CenterOnScreen()
	fa.win.Show()
	fa.win.RequestFocus()

	fa.reproducirAudio()

	inicio := time.Now()
	intervalo := 100 * time.Millisecond

	go func() {
		for fa.running {
			transcurrido := time.Since(inicio).Seconds()
			glowAlpha := uint8(math.Abs(math.Sin(transcurrido*5)) * 255)

			fyne.Do(func() {
				fa.fondo.FillColor = color.RGBA{R: 40 + glowAlpha/5, G: 10, B: 10, A: 255}
				fa.fondo.Refresh()
			})

			time.Sleep(intervalo)
		}
	}()
}

func (fa *FullscreenAlert) cerrar() {
	fa.running = false
	fa.detenerAudio()
	fa.win.Hide()
}

func (fa *FullscreenAlert) reproducirAudio() {
	cfg, err := fa.configMgr.GetConfig()
	if err != nil {
		return
	}
	deviceID := resolveAudioDeviceID(cfg.AudioDevice)
	go startAlertAudio(deviceID, fa.cancelAudio)
}

func (fa *FullscreenAlert) detenerAudio() {
	if fa.cancelAudio != nil {
		close(fa.cancelAudio)
		fa.cancelAudio = nil
	}
}
