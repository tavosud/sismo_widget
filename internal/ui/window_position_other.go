//go:build !windows

package ui

import "fyne.io/fyne/v2"

func SetWindowToTopRight(win fyne.Window, width, height float32) {
}

func SetFrameless(win fyne.Window) {
}

func StartNativeWindowDrag() {
}

func SetRoundedCorners() {
}

func GetWindowPosition() (int, int) {
	return 0, 0
}

func SetWindowPosition(x, y int) {
}

func IsAlreadyRunning() bool {
	return false
}
