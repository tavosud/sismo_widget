package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"net/http"
	"sort"
)

func tileXY(lat, lon float64, zoom int) (x, y int) {
	latRad := lat * math.Pi / 180
	n := math.Pow(2, float64(zoom))
	x = int(math.Floor((lon + 180) / 360 * n))
	y = int(math.Floor((1 - math.Log(math.Tan(latRad)+1/math.Cos(latRad))/math.Pi) / 2 * n))
	return
}

func pixelInTile(lat, lon float64, zoom, tileX, tileY int) (px, py int) {
	n := math.Pow(2, float64(zoom))
	worldX := (lon + 180) / 360 * n
	worldY := (1 - math.Log(math.Tan(lat*math.Pi/180)+1/math.Cos(lat*math.Pi/180))/math.Pi) / 2 * n
	px = int((worldX - float64(tileX)) * 256)
	py = int((worldY - float64(tileY)) * 256)
	return
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy
	for {
		if x0 >= img.Bounds().Min.X && x0 < img.Bounds().Max.X && y0 >= img.Bounds().Min.Y && y0 < img.Bounds().Max.Y {
			img.Set(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func drawPulseRing(img *image.RGBA, cx, cy, r int) {
	b := img.Bounds()
	for y := cy - r - 2; y <= cy+r+2; y++ {
		for x := cx - r - 2; x <= cx+r+2; x++ {
			d := math.Sqrt(float64((x-cx)*(x-cx) + (y-cy)*(y-cy)))
			if math.Abs(d-float64(r)) < 1.5 && x >= b.Min.X && x < b.Max.X && y >= b.Min.Y && y < b.Max.Y {
				alpha := uint8(200 - float64(r)*4)
				if alpha > 200 {
					alpha = 200
				}
				img.Set(x, y, color.RGBA{R: 255, G: 100, B: 50, A: alpha})
			}
		}
	}
}

func drawStar(img *image.RGBA, cx, cy int) {
	const outerR = 12
	const innerR = 5
	pts := make([]image.Point, 10)
	for i := 0; i < 10; i++ {
		angle := float64(i)*math.Pi/5 - math.Pi/2
		r := outerR
		if i%2 == 1 {
			r = innerR
		}
		pts[i].X = cx + int(float64(r)*math.Cos(angle)+0.5)
		pts[i].Y = cy + int(float64(r)*math.Sin(angle)+0.5)
	}
	b := img.Bounds()
	minY, maxY := pts[0].Y, pts[0].Y
	for _, p := range pts {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	starColor := color.RGBA{R: 220, G: 30, B: 30, A: 255}
	for y := minY; y <= maxY; y++ {
		var xs []int
		for i := 0; i < 10; i++ {
			j := (i + 1) % 10
			yi, yj := pts[i].Y, pts[j].Y
			if (yi <= y && yj > y) || (yj <= y && yi > y) {
				x := pts[i].X + (y-yi)*(pts[j].X-pts[i].X)/(yj-yi)
				xs = append(xs, x)
			}
		}
		sort.Ints(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			for x := xs[i]; x <= xs[i+1]; x++ {
				if x >= b.Min.X && x < b.Max.X && y >= b.Min.Y && y < b.Max.Y {
					img.Set(x, y, starColor)
				}
			}
		}
	}
	for i := 0; i < 10; i++ {
		j := (i + 1) % 10
		drawLine(img, pts[i].X, pts[i].Y, pts[j].X, pts[j].Y, color.White)
	}
}

func cargarTile(httpClient *http.Client, zoom, x, y int) (image.Image, error) {
	url := fmt.Sprintf("https://tile.openstreetmap.org/%d/%d/%d.png", zoom, x, y)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "SismoWidget/1.0 (IGP Monitor; sismowidget@local)")
	req.Header.Set("Accept", "image/png,image/*;q=0.9")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}
