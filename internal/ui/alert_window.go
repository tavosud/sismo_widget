package ui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"image/color"

	"sismo_widget/internal/entity"
)

type FullscreenAlert struct {
	app           fyne.App
	win           fyne.Window
	fondo         *canvas.Rectangle
	audioFile     *os.File
	audioStreamer beep.StreamSeeker
	speakerInit   bool
	running       bool
}

func NewFullscreenAlert(app fyne.App) *FullscreenAlert {
	alert := &FullscreenAlert{
		app: app,
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
	rutaAudio := filepath.Join("assets", "sounds", "alerta1.mp3")

	if _, err := os.Stat(rutaAudio); os.IsNotExist(err) {
		return
	}

	f, err := os.Open(rutaAudio)
	if err != nil {
		return
	}
	fa.audioFile = f

	streamer, _, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return
	}
	fa.audioStreamer = streamer

	if !fa.speakerInit {
		err = speaker.Init(44100, 44100/10)
		if err != nil {
			return
		}
		fa.speakerInit = true
	}

	speaker.Play(beep.Loop(-1, streamer))
}

func (fa *FullscreenAlert) detenerAudio() {
	speaker.Clear()
	if fa.audioFile != nil {
		fa.audioFile.Close()
		fa.audioFile = nil
	}
	fa.audioStreamer = nil
}
