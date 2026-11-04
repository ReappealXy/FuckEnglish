package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
)

var LibraryEvent = mygo.NewEvent[LibraryState]("library")

type Library struct{ core *Core }

func (l *Library) Snapshot() LibraryState { return l.core.snapshot() }
func (l *Library) Select(name string) (LibraryState, error) {
	return l.core.change(func() error {
		if _, ok := l.core.words[name]; !ok {
			return fmt.Errorf("词表不可用：%s，请刷新或检查文件格式", name)
		}
		l.core.cfg.Selected = name
		l.core.focus = 0
		return nil
	})
}
func (l *Library) Page(page int) (LibraryState, error) {
	return l.core.change(func() error {
		pages := (len(l.core.words[l.core.cfg.Selected]) + pageSize - 1) / pageSize
		l.core.cfg.Pages[l.core.cfg.Selected] = max(0, min(page, max(0, pages-1)))
		l.core.focus = 0
		return nil
	})
}
func (l *Library) Focus(index int) LibraryState {
	l.core.mu.Lock()
	l.core.focus = max(0, min(index, pageSize-1))
	s := l.core.snapshotLocked()
	l.core.mu.Unlock()
	_ = LibraryEvent.Broadcast(s)
	return s
}
func (l *Library) Refresh() (LibraryState, error) {
	return l.core.change(func() error { l.core.scanLocked(); return nil })
}
func (l *Library) SetFolder(folder string) (LibraryState, error) {
	abs, err := filepath.Abs(folder)
	if err != nil {
		return l.core.snapshot(), err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return l.core.snapshot(), fmt.Errorf("文件夹不可用：%w", err)
	}
	if !info.IsDir() {
		return l.core.snapshot(), fmt.Errorf("请选择文件夹")
	}
	return l.core.change(func() error {
		l.core.cfg.Folder = abs
		l.core.cfg.Selected = ""
		l.core.cfg.Pages = map[string]int{}
		l.core.scanLocked()
		if l.core.scanError != "" {
			return fmt.Errorf("%s", l.core.scanError)
		}
		return nil
	})
}

// PickFolder only picks. The user presses Save to switch the word directory.
func (l *Library) PickFolder(ctx context.Context) (string, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: mygo.CallerWindow(ctx), Title: "选择 Excel 词表文件夹", Directory: true, DefaultPath: l.core.snapshot().Folder})
	if err != nil || len(paths) == 0 {
		return "", err
	}
	return paths[0], nil
}
func (l *Library) OpenFolder() error { return mygo.Shell.OpenPath(l.core.snapshot().Folder) }

type Preferences struct{ core *Core }

func (p *Preferences) Get() Settings { return p.core.settings().Settings }
func (p *Preferences) Save(s Settings) (Settings, error) {
	valid := map[string]bool{"capsule": true, "notch": true, "line": true}
	if !valid[s.IslandStyle] {
		s.IslandStyle = "capsule"
	}
	s.IslandWidth = max(440, min(s.IslandWidth, 1200))
	s.IslandHeight = max(200, min(s.IslandHeight, 600))
	s.IslandExpandDelay = max(0, min(s.IslandExpandDelay, 1200))
	s.IslandCollapseDelay = max(120, min(s.IslandCollapseDelay, 2000))
	s.IslandAnimationDuration = max(100, min(s.IslandAnimationDuration, 1000))
	_, err := p.core.change(func() error { p.core.cfg.Settings = s; return nil })
	if err != nil {
		return p.Get(), err
	}
	_ = SettingsEvent.Broadcast(s)
	applyIsland(p.core)
	return s, nil
}
