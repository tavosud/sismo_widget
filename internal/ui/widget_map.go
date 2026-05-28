package ui

import (
	"context"
	"image"
	"image/draw"
	"math"
	"time"

	"fyne.io/fyne/v2"
)

func (w *widgetUI) cargarMapa(lat, lon float64) {
	if w.cancelPulse != nil {
		w.cancelPulse()
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancelPulse = cancel

	w.txtMapLoading.Text = "Cargando mapa..."
	w.txtMapLoading.Refresh()

	go func() {
		zoom := w.mapZoom
		tx, ty := tileXY(lat, lon, zoom)
		px, py := pixelInTile(lat, lon, zoom, tx, ty)

		var img1, img2 image.Image
		var err1, err2 error
		var xOff int

		if px < 128 && tx > 0 {
			img1, err1 = cargarTile(w.httpClient, zoom, tx-1, ty)
			img2, err2 = cargarTile(w.httpClient, zoom, tx, ty)
			xOff = 256
		} else {
			img1, err1 = cargarTile(w.httpClient, zoom, tx, ty)
			img2, err2 = cargarTile(w.httpClient, zoom, tx+1, ty)
			xOff = 0
		}
		if err1 != nil || err2 != nil {
			fyne.Do(func() {
				w.txtMapLoading.Text = "Mapa no disponible"
				w.txtMapLoading.Refresh()
			})
			return
		}

		rgba := image.NewRGBA(image.Rect(0, 0, 512, 256))
		draw.Draw(rgba, image.Rect(0, 0, 256, 256), img1, image.Point{}, draw.Over)
		draw.Draw(rgba, image.Rect(256, 0, 512, 256), img2, image.Point{}, draw.Over)

		cleanTile := image.NewRGBA(image.Rect(0, 0, 512, 256))
		draw.Draw(cleanTile, cleanTile.Bounds(), rgba, image.Point{}, draw.Over)

		sx, sy := px+xOff, py
		drawStar(rgba, sx, sy)

		fyne.Do(func() {
			w.mapTileImg.Image = rgba
			w.mapTileImg.Refresh()
			w.txtMapLoading.Text = ""
			w.txtMapLoading.Refresh()
		})

		go func(ctx context.Context, cx, cy int) {
			phase := 0.0
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(300 * time.Millisecond):
					phase += 0.2
					if phase > 1.0 {
						phase = 0.0
					}
					ringR := 12 + int(8*math.Abs(math.Sin(phase*math.Pi)))
					frame := image.NewRGBA(cleanTile.Bounds())
					draw.Draw(frame, frame.Bounds(), cleanTile, image.Point{}, draw.Over)
					drawPulseRing(frame, cx, cy, ringR)
					drawStar(frame, cx, cy)
					fyne.Do(func() {
						w.mapTileImg.Image = frame
						w.mapTileImg.Refresh()
					})
				}
			}
		}(ctx, sx, sy)
	}()
}
