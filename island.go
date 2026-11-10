package main

import (
	"log"
	"math"

	"github.com/egoist/mygo"
)

type IslandWindowState struct {
	Expanded bool `json:"expanded"`
}

// CSS reports the surface being drawn, including intermediate animation frames.
// The fixed host's remaining transparent area must not intercept other apps.
type IslandHitRegion struct {
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	TopRadius    float64 `json:"topRadius"`
	BottomRadius float64 `json:"bottomRadius"`
	Compact      bool    `json:"compact"`
}

func (r IslandHitRegion) valid() bool {
	for _, v := range []float64{r.X, r.Y, r.Width, r.Height, r.TopRadius, r.BottomRadius} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 16384 {
			return false
		}
	}
	return r.Width > 0 && r.Height > 0
}

func (r IslandHitRegion) contains(x, y float64) bool {
	x, y = x-r.X, y-r.Y
	if x < 0 || y < 0 || x >= r.Width || y >= r.Height {
		return false
	}
	top, bottom := min(r.TopRadius, r.Width/2, r.Height/2), min(r.BottomRadius, r.Width/2, r.Height/2)
	corner := top
	cy := top
	if y >= r.Height-bottom {
		corner, cy = bottom, r.Height-bottom
	} else if y >= top {
		return true
	}
	cx := corner
	if x >= r.Width-corner {
		cx = r.Width - corner
	} else if x >= corner {
		return true
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= corner*corner
}

var IslandWindowEvent = mygo.NewEvent[IslandWindowState]("island-window")
var IslandDismissEvent = mygo.NewEvent[bool]("island-dismiss")
var IslandPointerEvent = mygo.NewEvent[bool]("island-pointer")
var IslandPageEvent = mygo.NewEvent[int]("island-page")
var SettingsEvent = mygo.NewEvent[Settings]("settings")

type Island struct{ core *Core }

func (i *Island) OpenMain()                 { mygo.RunOnMain(showMainWindow) }
func (i *Island) SetExpanded(expanded bool) { setIslandExpanded(i.core, expanded) }
func (i *Island) SetPagingShortcuts(enabled bool) {
	mygo.RunOnMain(func() {
		islandPagingAllowed = enabled
		updateIslandPagingShortcuts()
	})
}
func (i *Island) SetHitRegion(region IslandHitRegion) {
	if !region.valid() {
		return
	}
	mygo.RunOnMain(func() { islandHitRegion = region })
}
func (i *Island) OpenSettings() {
	mygo.RunOnMain(func() {
		mainWindowURL = "/?page=island"
		showMainWindow()
		mainWindowURL = "/"
		_ = IslandNavigateEvent.Broadcast("island")
	})
}

var IslandNavigateEvent = mygo.NewEvent[string]("island-navigate")

// This controller is owned exclusively by the main thread.
var islandWin *mygo.Window
var islandReady bool
var islandExpanded bool
var mainWindowURL = "/"
var islandEscape bool
var islandHitRegion IslandHitRegion
var islandPointerInside bool
var islandPagingAllowed bool
var islandPagingActive bool
var islandPagingRegistered bool

func updateIslandPagingShortcuts() {
	active := !studyActive && islandExpanded && islandPointerInside && islandPagingAllowed
	if active == islandPagingActive {
		return
	}
	islandPagingActive = active
	if !active {
		if islandPagingRegistered {
			mygo.GlobalShortcut.Unregister("Left")
			mygo.GlobalShortcut.Unregister("Right")
			islandPagingRegistered = false
		}
		return
	}
	// Hovering does not activate the overlay, so DOM keydown alone cannot
	// receive these keys. Own them only while the pointer is over the open
	// browsing surface; native controls and quizzes keep their normal keys.
	if mygo.GlobalShortcut.IsRegistered("Left") || mygo.GlobalShortcut.IsRegistered("Right") {
		return
	}
	registered := []string{}
	for _, key := range []struct {
		name      string
		direction int
	}{{"Left", -1}, {"Right", 1}} {
		if err := mygo.GlobalShortcut.Register(key.name, func() {
			if !studyActive && islandExpanded && islandPointerInside && islandPagingAllowed {
				_ = IslandPageEvent.Broadcast(key.direction)
			}
		}); err != nil {
			for _, name := range registered {
				mygo.GlobalShortcut.Unregister(name)
			}
			log.Println("island paging shortcut:", err)
			return
		}
		registered = append(registered, key.name)
	}
	islandPagingRegistered = true
}

func islandEscapeKey(expanded bool) {
	if expanded && !islandEscape && !mygo.GlobalShortcut.IsRegistered("Escape") {
		islandEscape = mygo.GlobalShortcut.Register("Escape", func() {
			_ = IslandDismissEvent.Broadcast(true)
		}) == nil
	} else if !expanded && islandEscape {
		mygo.GlobalShortcut.Unregister("Escape")
		islandEscape = false
	}
}

func islandStyle(s Settings) string {
	switch s.IslandStyle {
	case "notch", "line", "capsule":
		return s.IslandStyle
	}
	return "capsule"
}

func islandDisplayArea(display mygo.Display, s Settings) mygo.Rectangle {
	// CollapsedGeometry.Resolve uses the physical screen for the notch and
	// the working area for the capsule and indicator line.
	if islandStyle(s) == "notch" {
		return display.Bounds
	}
	return display.WorkArea
}

func islandHostBounds(area mygo.Rectangle, s Settings) mygo.Rectangle {
	// BrimDeck applies its largest transparent layered host once (MainWindow
	// ApplySettings), then changes only Island inside that host. Resizing a
	// layered HWND while its contents animate makes Windows present a stale
	// opaque frame. Keep the extra 12/16-DIP room for the shadow and hover
	// area; CSS owns the compact 196x32 / 238x32 / 56x4 surface geometry.
	w, h := s.IslandWidth+24, s.IslandHeight+16
	if s.IslandWidth <= 0 {
		w = 544
	}
	if s.IslandHeight <= 0 {
		h = 216
	}
	w, h = min(w, area.Width), min(h, area.Height)
	return mygo.Rectangle{X: area.X + (area.Width-w)/2, Y: area.Y, Width: w, Height: h}
}

func applyIsland(core *Core) {
	mygo.RunOnMain(func() {
		if studyActive || !core.settings().IslandVisible {
			islandExpanded = false
			islandPointerInside, islandPagingAllowed = false, false
			updateIslandPagingShortcuts()
			islandEscapeKey(false)
			if islandWin != nil {
				islandWin.Hide()
				_ = IslandWindowEvent.Broadcast(IslandWindowState{Expanded: false})
			}
			return
		}
		if islandWin != nil && !islandWin.IsDestroyed() {
			s := core.settings().Settings
			islandWin.SetBounds(islandHostBounds(islandDisplayArea(mygo.Screen.PrimaryDisplay(), s), s))
			if islandReady {
				islandWin.ShowInactive()
			}
			return
		}
		s := core.settings().Settings
		b := islandHostBounds(islandDisplayArea(mygo.Screen.PrimaryDisplay(), s), s)
		w := mygo.NewWindow(islandWindowOptions(b))
		// Select the compositor-backed host at creation, then remove its material.
		// MyGo keeps WS_EX_NOREDIRECTIONBITMAP for this HWND's lifetime.
		w.SetVibrancy(mygo.VibrancyNone)
		if err := configureIslandOverlay(w); err != nil {
			log.Println("island overlay:", err)
		}
		islandWin, islandReady = w, false
		islandHitRegion = IslandHitRegion{}
		islandPointerInside, islandPagingAllowed = false, false
		updateIslandPagingShortcuts()
		stopPointer := startIslandPointer(w, core)
		w.OnClose(func(e *mygo.CloseEvent) {
			if applicationQuitting {
				return
			}
			e.PreventDefault()
			_ = IslandDismissEvent.Broadcast(true)
		})
		w.OnReadyToShow(func() {
			islandReady = true
			if !studyActive && core.settings().IslandVisible {
				w.ShowInactive()
			}
		})
		w.OnClosed(func() {
			stopPointer()
			islandWin, islandReady, islandExpanded = nil, false, false
			islandPointerInside, islandPagingAllowed = false, false
			updateIslandPagingShortcuts()
			islandEscapeKey(false)
		})
	})
}

func islandWindowOptions(b mygo.Rectangle) mygo.WindowOptions {
	return mygo.WindowOptions{
		Title: "FuckEnglish · 单词胶囊", URL: "/?view=island",
		X: b.X, Y: b.Y, Width: b.Width, Height: b.Height,
		Hidden: true, Frameless: true, Transparent: true,
		Vibrancy:        mygo.VibrancyHUD,
		BackgroundColor: "#00000000", AlwaysOnTop: true, SkipTaskbar: true,
		// MyGo v0.4.8's DisableMove blocks all WM_WINDOWPOSCHANGING moves,
		// including the initial placement and our own SetBounds calls.
		// Frameless content has no drag regions, so leave native moves enabled.
		DisableResize: true, DisableMinimize: true,
		DisableMaximize: true, DisableFullScreen: true, DisableShadow: true, DisableClose: true,
	}
}

func setIslandExpanded(core *Core, expanded bool) {
	mygo.RunOnMain(func() {
		if islandWin == nil || islandWin.IsDestroyed() {
			return
		}
		s := core.settings().Settings
		expanded = expanded && !studyActive && s.IslandVisible
		islandExpanded = expanded
		updateIslandPagingShortcuts()
		islandEscapeKey(expanded)
		if islandReady && !studyActive && core.settings().IslandVisible {
			islandWin.ShowInactive()
		}
		// The requesting view already owns the animation state. Do not echo
		// delayed close requests into a newer, reversed opening transition.
		// Forced hiding still broadcasts a reset from applyIsland above.
	})
}

func setupIsland(core *Core) {
	applyIsland(core)
	mygo.Screen.OnDisplaysChanged(func() {
		if islandWin != nil && !islandWin.IsDestroyed() {
			s := core.settings().Settings
			islandWin.SetBounds(islandHostBounds(islandDisplayArea(mygo.Screen.PrimaryDisplay(), s), s))
		}
	})
}
