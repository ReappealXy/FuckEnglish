package main

import "github.com/egoist/mygo"

// These window references and mode flags belong to the MyGo main thread.
var studyWin *mygo.Window
var studyReady bool
var studyActive bool

func (i *Island) OpenStudy() {
	mygo.RunOnMain(func() { showStudyWindow(i.core) })
}

func (i *Island) ReturnToCapsule() {
	mygo.RunOnMain(func() { returnToCapsule(i.core) })
}

func showStudyWindow(core *Core) {
	studyActive = true
	applyIsland(core)
	if studyWin != nil && !studyWin.IsDestroyed() {
		if studyReady {
			if studyWin.IsMinimized() {
				studyWin.Restore()
			}
			studyWin.Show()
			studyWin.Focus()
		}
		return
	}
	w := mygo.NewWindow(mygo.WindowOptions{
		Title: "FuckEnglish · 背单词", URL: "/?view=study",
		Width: 900, Height: 640, MinWidth: 560, MinHeight: 420,
		TitleBarStyle: mygo.TitleBarHidden, TitleBarHeight: 48,
		Transparent: true, BackgroundColor: "#00000000", Vibrancy: mygo.VibrancyHUD,
		Hidden: true, StateKey: "study",
	})
	// Keep a compositor-backed host so transparent CSS reveals the desktop,
	// then remove the system material to leave opacity under our two controls.
	w.SetVibrancy(mygo.VibrancyNone)
	studyWin, studyReady = w, false
	w.OnReadyToShow(func() {
		studyReady = true
		if studyActive && !applicationQuitting {
			w.Show()
			w.Focus()
		}
	})
	w.OnMinimize(func() {
		if !applicationQuitting {
			returnToCapsule(core)
		}
	})
	w.OnClose(func(e *mygo.CloseEvent) {
		if !applicationQuitting {
			e.PreventDefault()
			returnToCapsule(core)
		}
	})
	w.OnClosed(func() {
		studyWin, studyReady = nil, false
		if !applicationQuitting {
			returnToCapsule(core)
		}
	})
}

func returnToCapsule(core *Core) {
	studyActive = false
	if studyWin != nil && !studyWin.IsDestroyed() {
		studyWin.Hide()
	}
	applyIsland(core)
}
