//go:build windows

package main

import (
	"fmt"
	"unsafe"

	"github.com/egoist/mygo"
	"golang.org/x/sys/windows"
)

// The page paints its own outline in the background opacity group. Keep the
// native resize frame and caption controls, but suppress DWM's opaque outline
// and shadow, which do not inherit the page's transparency.
func configureStudyFrame(w *mygo.Window) error {
	policy := uint32(1) // DWMNCRP_DISABLED
	proc := windows.NewLazySystemDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
	hr, _, _ := proc.Call(w.NativeHandle(), windows.DWMWA_NCRENDERING_POLICY, uintptr(unsafe.Pointer(&policy)), unsafe.Sizeof(policy))
	if int32(hr) < 0 {
		return fmt.Errorf("cannot suppress system frame painting: HRESULT 0x%08x", uint32(hr))
	}
	return nil
}

// MyGo's SkipTaskbar uses ITaskbarList3.DeleteTab, which this Windows shell
// can immediately undo for an unowned normal window. Give only this resident
// study window a tool-window identity; keep its regular activation and frame.
func configureStudyTaskbar(w *mygo.Window) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	getStyle := user32.NewProc("GetWindowLongPtrW")
	setStyle := user32.NewProc("SetWindowLongPtrW")
	const toolWindow, appWindow = uintptr(0x80), uintptr(0x40000)
	index := ^uintptr(19) // GWL_EXSTYLE = -20
	handle := w.NativeHandle()
	style, _, _ := getStyle.Call(handle, index)
	wanted := (style | toolWindow) &^ appWindow
	setStyle.Call(handle, index, wanted)
	actual, _, err := getStyle.Call(handle, index)
	if actual != wanted {
		return fmt.Errorf("cannot hide study taskbar entry: %w", err)
	}
	user32.NewProc("SetWindowPos").Call(handle, 0, 0, 0, 0, 0, 0x0037) // FRAMECHANGED, NOMOVE, NOSIZE, NOZORDER, NOACTIVATE
	return nil
}
