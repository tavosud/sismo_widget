package main

import (
	_ "embed"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"sismo_widget/internal/repository"
	"sismo_widget/internal/ui"
	"sismo_widget/internal/usecase"
)

//go:embed logo.png
var logoPNG []byte

func main() {
	if ui.IsAlreadyRunning() {
		return
	}
	configPath, historyPath, err := ui.ObtenerPathsConfig()
	if err != nil {
		log.Fatalf("Error al obtener rutas de configuración: %v", err)
	}

	configRepo := repository.NewConfigFileRepository(configPath)
	historyRepo := repository.NewHistoryFileRepository(historyPath)
	igpRepo := repository.NewIGPApiRepository()

	config, err := configRepo.Load()
	if err != nil {
		log.Printf("Advertencia: no se pudo cargar la configuración, usando valores predeterminados: %v", err)
	}

	configManager := usecase.NewConfigManager(configRepo)
	monitor := usecase.NewEarthquakeMonitor(igpRepo, historyRepo, config)

	myApp := app.NewWithID("com.sismowidget.app")

	if len(logoPNG) > 0 {
		myApp.SetIcon(&fyne.StaticResource{
			StaticName:    "logo.png",
			StaticContent: logoPNG,
		})
	} else {
		myApp.SetIcon(theme.WarningIcon())
	}

	if _, err := os.Stat(configPath); err == nil {
		ui.MostrarWidget(myApp, configManager, monitor, false)
	} else {
		ui.MostrarVentanaConfig(myApp, configManager, monitor, nil, true)
	}
	myApp.Run()
}
