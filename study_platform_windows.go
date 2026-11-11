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
