// FuckEnglish is a local vocabulary capsule derived from the existing island host.
package main

import (
	"github.com/egoist/mygo"
	"log"
	"os"
)

var applicationQuitting bool
var mainWin *mygo.Window
var appCore *Core

func main() {
	app := mygo.App
	app.SetName("FuckEnglish")
	core := &Core{}
	appCore = core
	mygo.Bind(&Library{core})
	mygo.Bind(&Preferences{core})
	mygo.Bind(&Island{core})
	if !app.RequestSingleInstanceLock() {
		return
	}
	app.OnSecondInstance(func([]string, string) { showMainWindow() })
	app.OnBeforeQuit(func(*mygo.QuitEvent) { applicationQuitting = true })
	app.OnWindowAllClosed(func() {})
	app.WhenReady(func() {
		dir, err := app.Path(mygo.PathUserData)
		if override := os.Getenv("FUCKENGLISH_DATA_DIR"); override != "" {
			dir = override
			err = nil
		}
		if err == nil {
			err = core.Init(dir)
		}
		if err != nil {
			mygo.Dialog.Error("FuckEnglish 无法启动", err.Error())
			app.Quit()
			return
		}
		setupTray(core)
		setupIsland(core)
		if !app.WasOpenedAtLogin() {
			showMainWindow()
		}
	})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
func showMainWindow() {
	if mainWin != nil && !mainWin.IsDestroyed() {
		if mainWin.IsMinimized() {
			mainWin.Restore()
		}
		mainWin.Show()
		mainWin.Focus()
		return
	}
	w := mygo.NewWindow(mygo.WindowOptions{Title: "FuckEnglish", Width: 1020, Height: 760, MinWidth: 840, MinHeight: 620, TitleBarStyle: mygo.TitleBarHidden, TitleBarHeight: 48, BackgroundColor: "light-dark(#f6f7f9, #111318)", StateKey: "main", URL: mainWindowURL})
	w.OnClose(func(e *mygo.CloseEvent) {
		if !applicationQuitting {
			e.PreventDefault()
			w.Hide()
		}
	})
	w.OnClosed(func() { mainWin = nil })
	mainWin = w
}
func setupTray(core *Core) {
	_, err := mygo.NewTray(mygo.TrayOptions{Icon: trayIcon, ToolTip: "FuckEnglish · 摸鱼背单词", Menu: mygo.NewMenu([]*mygo.MenuItem{
		{Label: "打开 FuckEnglish", Click: func(*mygo.MenuItem, *mygo.Window) { showMainWindow() }},
		{Label: "显示 / 隐藏胶囊", Click: func(*mygo.MenuItem, *mygo.Window) {
			go func() {
				p := &Preferences{core}
				s := p.Get()
				s.IslandVisible = !s.IslandVisible
				if _, e := p.Save(s); e != nil {
					log.Println(e)
				}
			}()
		}},
		{Label: "刷新词表", Click: func(*mygo.MenuItem, *mygo.Window) {
			go func() {
				_, err := (&Library{core}).Refresh()
				if err != nil {
					log.Println(err)
				}
			}()
		}},
		mygo.Separator(), {Label: "退出", Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	})})
	if err != nil {
		log.Println("tray:", err)
	}
}
