package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"fuckenglish/internal/wordbook"
)

//go:embed words/*.xlsx
var examples embed.FS

const pageSize = 5

type Settings struct {
	IslandVisible           bool   `json:"islandVisible"`
	IslandStyle             string `json:"islandStyle"`
	IslandTheme             string `json:"islandTheme"`
	IslandHover             bool   `json:"islandHover"`
	IslandExpandDelay       int    `json:"islandExpandDelay"`
	IslandCollapseDelay     int    `json:"islandCollapseDelay"`
	IslandAnimationDuration int    `json:"islandAnimationDuration"`
	IslandWidth             int    `json:"islandWidth"`
	IslandHeight            int    `json:"islandHeight"`
}
type settingsFile struct{ Settings }
type config struct {
	Settings
	Folder   string         `json:"folder"`
	Selected string         `json:"selected"`
	Pages    map[string]int `json:"pages"`
	PageSize int            `json:"pageSize"`
}
type Book struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Error string `json:"error"`
}
type LibraryState struct {
	Folder   string          `json:"folder"`
	Books    []Book          `json:"books"`
	Selected string          `json:"selected"`
	Words    []wordbook.Word `json:"words"`
	Page     int             `json:"page"`
	Pages    int             `json:"pages"`
	Total    int             `json:"total"`
	Focus    int             `json:"focus"`
	Error    string          `json:"error"`
}
type Core struct {
	mu        sync.Mutex
	path      string
	cfg       config
	books     []Book
	words     map[string][]wordbook.Word
	focus     int
	scanError string
}

func defaultSettings() Settings {
	return Settings{true, "capsule", "dark", true, 160, 420, 360, 520, 200}
}
func (c *Core) Init(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	c.path = filepath.Join(dir, "settings.json")
	c.cfg = config{Settings: defaultSettings(), Pages: map[string]int{}, PageSize: pageSize}
	b, err := os.ReadFile(c.path)
	first := errors.Is(err, os.ErrNotExist)
	if err != nil && !first {
		return err
	}
	if !first {
		if err := json.Unmarshal(b, &c.cfg); err != nil {
			return fmt.Errorf("设置文件损坏，请备份后移走 %s：%w", c.path, err)
		}
	}
	if c.cfg.Pages == nil {
		c.cfg.Pages = map[string]int{}
	}
	if c.cfg.PageSize > 0 && c.cfg.PageSize != pageSize {
		for book, page := range c.cfg.Pages {
			c.cfg.Pages[book] = max(0, page*c.cfg.PageSize/pageSize)
		}
	}
	c.cfg.PageSize = pageSize
	if c.cfg.Folder == "" {
		exe, _ := os.Executable()
		c.cfg.Folder = filepath.Join(filepath.Dir(exe), "words")
		if err := os.MkdirAll(c.cfg.Folder, 0755); err != nil {
			c.cfg.Folder = filepath.Join(dir, "words")
		}
	}
	if first {
		if err := os.MkdirAll(c.cfg.Folder, 0755); err != nil {
			return err
		}
		entries, _ := examples.ReadDir("words")
		for _, entry := range entries {
			name := filepath.Join(c.cfg.Folder, entry.Name())
			if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
				data, _ := examples.ReadFile("words/" + entry.Name())
				if err := os.WriteFile(name, data, 0644); err != nil {
					return err
				}
			}
		}
	}
	c.scanLocked()
	return c.saveLocked()
}
func (c *Core) settings() settingsFile {
	c.mu.Lock()
	defer c.mu.Unlock()
	return settingsFile{c.cfg.Settings}
}
func (c *Core) saveLocked() error {
	b, err := json.MarshalIndent(c.cfg, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(c.path), "settings-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, c.path)
}
func (c *Core) scanLocked() {
	c.books = []Book{}
	c.words = map[string][]wordbook.Word{}
	c.scanError = ""
	entries, err := os.ReadDir(c.cfg.Folder)
	if err != nil {
		c.scanError = "无法读取词表文件夹：" + err.Error()
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), "~$") || !strings.EqualFold(filepath.Ext(entry.Name()), ".xlsx") {
			continue
		}
		words, e := wordbook.Read(filepath.Join(c.cfg.Folder, entry.Name()))
		book := Book{Name: entry.Name(), Count: len(words)}
		if e != nil {
			book.Error = e.Error()
		} else {
			c.words[entry.Name()] = words
		}
		c.books = append(c.books, book)
	}
	sort.Slice(c.books, func(i, j int) bool { return c.books[i].Name < c.books[j].Name })
	if _, ok := c.words[c.cfg.Selected]; !ok {
		c.cfg.Selected = ""
		for _, b := range c.books {
			if b.Error == "" {
				c.cfg.Selected = b.Name
				break
			}
		}
	}
	c.focus = 0
}
func (c *Core) snapshotLocked() LibraryState {
	words := c.words[c.cfg.Selected]
	pages := (len(words) + pageSize - 1) / pageSize
	page := max(0, min(c.cfg.Pages[c.cfg.Selected], max(0, pages-1)))
	start := min(page*pageSize, len(words))
	end := min(start+pageSize, len(words))
	visible := append([]wordbook.Word{}, words[start:end]...)
	return LibraryState{c.cfg.Folder, c.books, c.cfg.Selected, visible, page, pages, len(words), max(0, min(c.focus, max(0, len(visible)-1))), c.scanError}
}
func (c *Core) snapshot() LibraryState { c.mu.Lock(); defer c.mu.Unlock(); return c.snapshotLocked() }
func (c *Core) publish()               { _ = LibraryEvent.Broadcast(c.snapshot()) }
func (c *Core) change(run func() error) (LibraryState, error) {
	c.mu.Lock()
	// Preferences and navigation are committed together; restore on disk failures.
	old, _ := json.Marshal(c.cfg)
	err := run()
	if err == nil {
		err = c.saveLocked()
	}
	if err != nil {
		var restored config
		_ = json.Unmarshal(old, &restored)
		c.cfg = restored
		c.scanLocked()
	}
	s := c.snapshotLocked()
	c.mu.Unlock()
	if err == nil {
		_ = LibraryEvent.Broadcast(s)
	}
	return s, err
}
