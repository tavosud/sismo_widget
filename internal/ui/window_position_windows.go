//go:build windows

package ui

import (
	"syscall"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	enumWindows        = user32.NewProc("EnumWindows")
	getWindowText      = user32.NewProc("GetWindowTextW")
	setWindowPos       = user32.NewProc("SetWindowPos")
	getSystemMetrics   = user32.NewProc("GetSystemMetrics")
	getWindowLong      = user32.NewProc("GetWindowLongW")
	setWindowLong      = user32.NewProc("SetWindowLongW")
	getWindowRect      = user32.NewProc("GetWindowRect")
	showWindow         = user32.NewProc("ShowWindow")
	releaseCapture     = user32.NewProc("ReleaseCapture")
	sendMessage        = user32.NewProc("SendMessageW")
	dwmapi             = syscall.NewLazyDLL("dwmapi.dll")
	dwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	createMutex        = kernel32.NewProc("CreateMutexW")
)

const (
	SM_CXSCREEN      = 0
	SM_CYSCREEN      = 1
	SWP_NOZORDER     = 0x0004
	SWP_NOACTIVATE   = 0x0010
	SWP_NOMOVE       = 0x0002
	SWP_NOSIZE       = 0x0001
	SWP_FRAMECHANGED = 0x0020
	GWL_STYLE        = ^uintptr(15) // -16 as uintptr
	WS_POPUP         = 0x80000000
	WS_CAPTION       = 0x00C00000
	WS_THICKFRAME    = 0x00040000
	WS_MINIMIZEBOX   = 0x00020000
	WS_MAXIMIZEBOX   = 0x00010000
	WS_SYSMENU       = 0x00080000
	SW_SHOW          = 5
	WM_NCLBUTTONDOWN = 0xA1
	HTCAPTION        = 2
)

var cachedHwnd uintptr

func cacheWindowHandle(title string) uintptr {
	if cachedHwnd != 0 {
		return cachedHwnd
	}
	cachedHwnd = findWindowByTitle(title)
	return cachedHwnd
}

func findWindowByTitle(title string) uintptr {
	var hwnd uintptr
	cb := syscall.NewCallback(func(h uintptr, lParam uintptr) uintptr {
		var buf [256]uint16
		getWindowText.Call(h, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if syscall.UTF16ToString(buf[:]) == title {
			hwnd = h
			return 0
		}
		return 1
	})
	enumWindows.Call(cb, 0)
	return hwnd
}

func SetWindowToTopRight(win fyne.Window, width, height float32) {
	screenWidth, _, _ := getSystemMetrics.Call(uintptr(SM_CXSCREEN))

	x := int(screenWidth) - int(width) - 20
	y := 20

	time.Sleep(100 * time.Millisecond)

	targetTitle := "IGP Sismo Monitor"
	if win.Title() != "" {
		targetTitle = win.Title()
	}

	hwnd := cacheWindowHandle(targetTitle)
	if hwnd != 0 {
		setWindowPos.Call(
			hwnd,
			0,
			uintptr(x),
			uintptr(y),
			uintptr(width),
			uintptr(height),
			SWP_NOZORDER|SWP_NOACTIVATE,
		)
	}
}

func SetFrameless(win fyne.Window) {
	targetTitle := "IGP Sismo Monitor"
	if win.Title() != "" {
		targetTitle = win.Title()
	}

	hwnd := cacheWindowHandle(targetTitle)
	if hwnd == 0 {
		return
	}

	style, _, _ := getWindowLong.Call(hwnd, GWL_STYLE)
	style &^= WS_CAPTION | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX | WS_SYSMENU
	style |= WS_POPUP
	setWindowLong.Call(hwnd, GWL_STYLE, style)
	setWindowPos.Call(hwnd, 0, 0, 0, 0, 0,
		SWP_FRAMECHANGED|SWP_NOMOVE|SWP_NOSIZE|SWP_NOZORDER)
	showWindow.Call(hwnd, SW_SHOW)
}

func StartNativeWindowDrag() {
	if cachedHwnd == 0 {
		return
	}
	releaseCapture.Call()
	sendMessage.Call(cachedHwnd, WM_NCLBUTTONDOWN, HTCAPTION, 0)
}

func SetRoundedCorners() {
	if cachedHwnd == 0 {
		return
	}
	pref := uint32(2) // DWMWCP_ROUND
	dwmSetWindowAttribute.Call(
		cachedHwnd,
		33, // DWMWA_WINDOW_CORNER_PREFERENCE
		uintptr(unsafe.Pointer(&pref)),
		unsafe.Sizeof(pref),
	)
}

func GetWindowPosition() (int, int) {
	if cachedHwnd == 0 {
		return 0, 0
	}
	var rect struct{ Left, Top, Right, Bottom int32 }
	getWindowRect.Call(cachedHwnd, uintptr(unsafe.Pointer(&rect)))
	return int(rect.Left), int(rect.Top)
}

func SetWindowPosition(x, y int) {
	if cachedHwnd == 0 {
		return
	}
	setWindowPos.Call(
		cachedHwnd,
		0,
		uintptr(x),
		uintptr(y),
		0, 0,
		SWP_NOZORDER|SWP_NOSIZE,
	)
}

func IsAlreadyRunning() bool {
	name, _ := syscall.UTF16PtrFromString("SismoWidgetSingleInstance")
	handle, _, err := createMutex.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		return false
	}
	return err == syscall.ERROR_ALREADY_EXISTS
}
