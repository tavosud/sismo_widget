package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"sismo_widget/internal/repository"
	"sismo_widget/internal/ui"
	"sismo_widget/internal/usecase"
)

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

	iconPath := filepath.Join("assets", "images", "logo.png")
	if _, err := os.Stat(iconPath); err == nil {
		iconBytes, _ := os.ReadFile(iconPath)
		myApp.SetIcon(&fyne.StaticResource{
			StaticName:    "logo.png",
			StaticContent: iconBytes,
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
