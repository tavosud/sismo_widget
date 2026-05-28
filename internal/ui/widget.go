package ui

import (
	"context"
	"image/color"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"sismo_widget/internal/usecase"
)

type windowDragOverlay struct {
	canvas.Rectangle
	dragStarted bool
	onDragEnd   func()
}

func (w *windowDragOverlay) Dragged(e *fyne.DragEvent) {
	if !w.dragStarted {
		w.dragStarted = true
		StartNativeWindowDrag()
	}
}

func (w *windowDragOverlay) DragEnd() {
	w.dragStarted = false
	if w.onDragEnd != nil {
		w.onDragEnd()
	}
}

type tooltipButton struct {
	widget.Button
	tooltip string
}

func newTooltipButton(label string, icon fyne.Resource, tooltip string, tapped func()) *tooltipButton {
	b := &tooltipButton{tooltip: tooltip, Button: widget.Button{Text: label, Icon: icon, OnTapped: tapped}}
	b.ExtendBaseWidget(b)
	return b
}

func (b *tooltipButton) ToolTipText() string {
	return b.tooltip
}

var widgetWindow fyne.Window

type widgetUI struct {
	myWindow  fyne.Window
	fyneApp   fyne.App
	configMgr *usecase.ConfigManager
	monitor   *usecase.EarthquakeMonitor

	mapZoom     int
	mapLat      float64
	mapLon      float64
	cancelPulse context.CancelFunc

	clrs struct {
		verde    color.Color
		amarillo color.Color
		rojo     color.Color
		fondo    color.Color
		negro    color.Color
	}

	headerBg       *canvas.Rectangle
	txtMagnitud    *canvas.Text
	txtUbicacion   *canvas.Text
	txtFechaHeader *canvas.Text

	mapTileImg    *canvas.Image
	txtMapLoading *canvas.Text
	httpClient    *http.Client

	txtProfundidad *canvas.Text
	txtIntensidad  *canvas.Text
	txtDistancia   *canvas.Text
}

func (w *widgetUI) initColors() {
	w.clrs.verde = color.RGBA{R: 0, G: 128, B: 0, A: 255}
	w.clrs.amarillo = color.RGBA{R: 230, G: 200, B: 0, A: 255}
	w.clrs.rojo = color.RGBA{R: 200, G: 30, B: 30, A: 255}
	w.clrs.fondo = color.RGBA{R: 248, G: 248, B: 248, A: 255}
	w.clrs.negro = color.RGBA{R: 0, G: 0, B: 0, A: 255}
}

func MostrarWidget(
	fyneApp fyne.App,
	configManager *usecase.ConfigManager,
	monitor *usecase.EarthquakeMonitor,
	isSimulacion bool,
) {
	if widgetWindow != nil {
		widgetWindow.Show()
		if isSimulacion {
			simularAlerta(fyneApp, configManager, monitor)
		}
		return
	}

	w := &widgetUI{
		fyneApp:    fyneApp,
		configMgr:  configManager,
		monitor:    monitor,
		mapZoom:    10,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	w.initColors()

	w.myWindow = fyneApp.NewWindow("IGP Sismo Monitor")
	widgetWindow = w.myWindow
	if isSimulacion {
		w.myWindow.SetTitle("IGP Sismo Monitor - SIMULACRO")
	}

	winSize := fyne.NewSize(500, 410)
	w.myWindow.Resize(winSize)
	w.myWindow.SetPadded(false)

	w.myWindow.SetCloseIntercept(func() {
		guardarPosicionVentana(configManager)
		w.myWindow.Hide()
		iniciarSystray(fyneApp, w.myWindow, configManager)
	})

	w.myWindow.Show()
	SetFrameless(w.myWindow)
	SetRoundedCorners()
	cfg, _ := configManager.GetConfig()
	if cfg.WindowX != 0 || cfg.WindowY != 0 {
		SetWindowPosition(cfg.WindowX, cfg.WindowY)
	} else {
		SetWindowToTopRight(w.myWindow, winSize.Width, winSize.Height)
	}

	header := w.buildHeader()
	mapSection := w.buildMapSection()
	dataSection := w.buildDataSection()
	bottomBar := w.buildBottomBar()

	content := container.NewVBox(mapSection, dataSection)
	scrollContent := container.NewVScroll(container.NewPadded(content))
	fondoGeneral := canvas.NewRectangle(w.clrs.fondo)

	w.myWindow.SetContent(container.NewStack(fondoGeneral,
		container.NewBorder(header, bottomBar, nil, nil, scrollContent)))

	if isSimulacion {
		w.handleSimulation()
	} else {
		w.startPolling()
	}
}

func (w *widgetUI) buildHeader() fyne.CanvasObject {
	w.headerBg = canvas.NewRectangle(w.clrs.verde)
	w.txtMagnitud = canvas.NewText("M --", color.White)
	w.txtMagnitud.TextSize = 32
	w.txtMagnitud.TextStyle = fyne.TextStyle{Bold: true}
	w.txtUbicacion = canvas.NewText("Cargando datos...", color.White)
	w.txtUbicacion.TextSize = 13
	w.txtFechaHeader = canvas.NewText("--", color.White)
	w.txtFechaHeader.TextSize = 11

	btnCompartir := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		url := "https://ultimosismo.igp.gob.pe/ultimo-sismo"
		if clip := w.myWindow.Clipboard(); clip != nil {
			clip.SetContent(url)
		}
		w.fyneApp.SendNotification(&fyne.Notification{
			Title:   "Compartir",
			Content: "Enlace copiado al portapapeles",
		})
	})

	headerDrag := &windowDragOverlay{
		Rectangle: *canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 0}),
		onDragEnd: func() { guardarPosicionVentana(w.configMgr) },
	}

	headerLeft := container.NewVBox(w.txtMagnitud, w.txtUbicacion, w.txtFechaHeader)
	headerDragStack := container.NewStack(headerLeft, headerDrag)
	headerContent := container.NewBorder(nil, nil, headerDragStack, container.NewPadded(btnCompartir), nil)
	return container.NewStack(w.headerBg, container.NewPadded(headerContent))
}

func (w *widgetUI) buildMapSection() fyne.CanvasObject {
	w.mapTileImg = canvas.NewImageFromImage(nil)
	w.mapTileImg.FillMode = canvas.ImageFillStretch
	w.mapTileImg.SetMinSize(fyne.NewSize(460, 260))

	w.txtMapLoading = canvas.NewText("Cargando mapa...", color.RGBA{R: 150, G: 150, B: 150, A: 255})
	w.txtMapLoading.Alignment = fyne.TextAlignCenter

	mapBg := canvas.NewRectangle(color.RGBA{R: 234, G: 239, B: 247, A: 255})
	mapBorder := canvas.NewRectangle(color.RGBA{R: 200, G: 200, B: 200, A: 255})

	btnZoomIn := widget.NewButton("+", func() {
		if w.mapZoom < 18 {
			w.mapZoom++
			w.cargarMapa(w.mapLat, w.mapLon)
		}
	})
	btnZoomOut := widget.NewButton("-", func() {
		if w.mapZoom > 3 {
			w.mapZoom--
			w.cargarMapa(w.mapLat, w.mapLon)
		}
	})
	zoomControls := container.NewVBox(btnZoomOut, btnZoomIn)

	mapStack := container.NewStack(mapBg, container.NewCenter(w.txtMapLoading), w.mapTileImg)
	mapWithZoom := container.NewStack(mapStack,
		container.NewBorder(nil, nil, nil, container.NewPadded(zoomControls), nil))
	return container.NewPadded(container.NewStack(mapBorder, container.NewPadded(mapWithZoom)))
}

func (w *widgetUI) buildDataSection() fyne.CanvasObject {
	lblProfundidad := canvas.NewText("⬇️ Profundidad", w.clrs.negro)
	lblProfundidad.TextStyle = fyne.TextStyle{Bold: true}
	w.txtProfundidad = canvas.NewText("-- km", w.clrs.negro)

	lblIntensidad := canvas.NewText("📡 Intensidad", w.clrs.negro)
	lblIntensidad.TextStyle = fyne.TextStyle{Bold: true}
	w.txtIntensidad = canvas.NewText("--", w.clrs.negro)

	lblDistancia := canvas.NewText("📏 Distancia", w.clrs.negro)
	lblDistancia.TextStyle = fyne.TextStyle{Bold: true}
	w.txtDistancia = canvas.NewText("-- km", w.clrs.negro)

	return container.NewVBox(container.NewGridWithColumns(3,
		container.NewVBox(container.NewCenter(lblProfundidad), container.NewCenter(w.txtProfundidad)),
		container.NewVBox(container.NewCenter(lblIntensidad), container.NewCenter(w.txtIntensidad)),
		container.NewVBox(container.NewCenter(lblDistancia), container.NewCenter(w.txtDistancia)),
	))
}

func (w *widgetUI) buildBottomBar() fyne.CanvasObject {
	btnConfig := newTooltipButton("Configuración", theme.SettingsIcon(), "Configuración", func() {
		guardarPosicionVentana(w.configMgr)
		w.myWindow.Hide()
		MostrarVentanaConfig(w.fyneApp, w.configMgr, w.monitor, widgetWindow, false)
	})
	btnHistorial := newTooltipButton("Historial", theme.ListIcon(), "Historial", func() {
		MostrarVentanaHistorial(w.fyneApp, w.monitor)
	})
	btnCerrar := newTooltipButton("", theme.WindowCloseIcon(), "Cerrar", func() {
		guardarPosicionVentana(w.configMgr)
		w.myWindow.Hide()
		iniciarSystray(w.fyneApp, w.myWindow, w.configMgr)
	})
	return container.NewPadded(container.NewGridWithColumns(3, btnConfig, btnHistorial, btnCerrar))
}
