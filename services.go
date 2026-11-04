package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"fuckenglish/internal/wordbook"
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
		direction := l.core.quiz.direction
		changed := l.core.cfg.Selected != name
		l.core.cfg.Selected = name
		if changed && direction != "" {
			if err := l.core.startQuizLocked(direction); err != nil {
				return err
			}
		}
		l.core.focus = 0
		return nil
	})
}
func (l *Library) Page(page int) (LibraryState, error) {
	return l.core.change(func() error {
		pages := (len(l.core.words[l.core.cfg.Selected]) + pageSize - 1) / pageSize
		l.core.browsePagesLocked()[l.core.cfg.Selected] = max(0, min(page, max(0, pages-1)))
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

func (l *Library) SetBrowseOrder(order string) (LibraryState, error) {
	return l.core.change(func() error {
		if order != "sequential" && order != "random" {
			return fmt.Errorf("请选择顺序或乱序")
		}
		l.core.cfg.BrowseOrder = order
		if order == "random" && l.core.cfg.ShuffleSeed == 0 {
			l.core.cfg.ShuffleSeed = rand.Uint64()
		}
		l.core.focus = 0
		return nil
	})
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
		l.core.cfg.RandomPages = map[string]int{}
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

func (l *Library) SaveTemplate(ctx context.Context) (string, error) {
	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      mygo.CallerWindow(ctx),
		Title:       "保存 Excel 单词模板",
		DefaultPath: "FuckEnglish-单词模板.xlsx",
		Filters:     []mygo.FileFilter{{Name: "Excel 工作簿", Extensions: []string{"xlsx"}}},
	})
	if err != nil || path == "" {
		return "", err
	}
	if !strings.EqualFold(filepath.Ext(path), ".xlsx") {
		path += ".xlsx"
	}
	if err := wordbook.WriteTemplate(path); err != nil {
		return "", err
	}
	return path, nil
}

type Preferences struct{ core *Core }

func (p *Preferences) Get() Settings {
	s := p.core.settings().Settings
	s.OpenAtLogin = mygo.App.OpenAtLogin()
	return s
}
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
	wasOpenAtLogin := mygo.App.OpenAtLogin()
	if s.OpenAtLogin != wasOpenAtLogin {
		if err := mygo.App.SetOpenAtLogin(s.OpenAtLogin); err != nil {
			return p.Get(), err
		}
	}
	_, err := p.core.change(func() error { p.core.cfg.Settings = s; return nil })
	if err != nil {
		if s.OpenAtLogin != wasOpenAtLogin {
			if restoreErr := mygo.App.SetOpenAtLogin(wasOpenAtLogin); restoreErr != nil {
				return p.Get(), fmt.Errorf("%w; 恢复开机启动状态失败：%v", err, restoreErr)
			}
		}
		return p.Get(), err
	}
	_ = SettingsEvent.Broadcast(s)
	applyIsland(p.core)
	return s, nil
}
