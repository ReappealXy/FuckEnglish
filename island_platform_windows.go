//go:build windows

package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/egoist/mygo"
	"golang.org/x/sys/windows"
)

// BrimDeck's overlay does not activate on hover. Tool-window identity also
// keeps this app-owned surface out of Alt+Tab and taskbar window previews.
func configureIslandOverlay(w *mygo.Window) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	getStyle := user32.NewProc("GetWindowLongPtrW")
	setStyle := user32.NewProc("SetWindowLongPtrW")
	const noActivate, toolWindow, appWindow = uintptr(0x08000000), uintptr(0x80), uintptr(0x40000)
	// GWL_EXSTYLE = -20, represented as an unsigned LPARAM index.
	index := ^uintptr(19)
	handle := w.NativeHandle()
	// MyGo's frameless WebView keeps WS_OVERLAPPEDWINDOW and only removes
	// the non-client layout. An overlay must also remove the DWM frame,
	// caption buttons and system shadow, like BrimDeck's WindowStyle=None.
	styleIndex := ^uintptr(15) // GWL_STYLE = -16
	base, _, _ := getStyle.Call(handle, styleIndex)
	setStyle.Call(handle, styleIndex, (base&^uintptr(0x00CF0000))|uintptr(0x80000000))
	user32.NewProc("SetWindowPos").Call(handle, 0, 0, 0, 0, 0, 0x0037) // FRAMECHANGED, NOMOVE, NOSIZE, NOZORDER, NOACTIVATE
	style, _, _ := getStyle.Call(handle, index)
	wanted := (style | noActivate | toolWindow) &^ appWindow
	setStyle.Call(handle, index, wanted)
	actual, _, err := getStyle.Call(handle, index)
	if actual != wanted {
		return fmt.Errorf("could not apply resident overlay style: %w", err)
	}
	// Transparent WebView pixels alone reveal MyGo's opaque black GDI brush.
	// Extend the transparent DWM frame over the entire fixed host, not the
	// one-pixel shadow frame (DisableShadow intentionally removes that).
	dwm := windows.NewLazySystemDLL("dwmapi.dll")
	margins := [4]int32{-1, -1, -1, -1}
	hr, _, _ := dwm.NewProc("DwmExtendFrameIntoClientArea").Call(handle, uintptr(unsafe.Pointer(&margins)))
	if int32(hr) < 0 {
		return fmt.Errorf("could not make island host transparent: HRESULT 0x%08x", hr)
	}
	return nil
}

// BrimDeck MainWindow.TrackPointer polls the real cursor, rather than painting
// a large invisible hover panel. This also works when the host ignores input.
func startIslandPointer(w *mygo.Window, core *Core) func() {
	user32 := windows.NewLazySystemDLL("user32.dll")
	getCursor, toClient, getDPI := user32.NewProc("GetCursorPos"), user32.NewProc("ScreenToClient"), user32.NewProc("GetDpiForWindow")
	stop := make(chan struct{})
	inside, ignoring := false, false
	go func() {
		ticker := time.NewTicker(32 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				mygo.RunOnMain(func() {
					if w.IsDestroyed() || !islandReady || !core.settings().IslandVisible {
						return
					}
					var point [2]int32
					if ok, _, _ := getCursor.Call(uintptr(unsafe.Pointer(&point))); ok == 0 {
						return
					}
					toClient.Call(w.NativeHandle(), uintptr(unsafe.Pointer(&point)))
					dpi, _, _ := getDPI.Call(w.NativeHandle())
					if dpi == 0 {
						dpi = 96
					}
					x, y := float64(point[0])*96/float64(dpi), float64(point[1])*96/float64(dpi)
					r := islandHitRegion
					hit := r.valid() && r.contains(x, y)
					if ignoring == hit {
						ignoring = !hit
						w.SetIgnoreMouseEvents(ignoring)
					}
					// The compact sensing box bridges its five-DIP top gap, as
					// CollapsedHover does, without drawing or blocking that gap.
					hover := hit || (r.valid() && r.Compact && x >= r.X && x < r.X+r.Width && y >= 0 && y < r.Y+r.Height)
					if hover != inside {
						inside = hover
						_ = IslandPointerEvent.Broadcast(inside)
					}
				})
			}
		}
	}()
	return func() { close(stop) }
}
