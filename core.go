package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
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
	OpenAtLogin             bool   `json:"openAtLogin"`
}
type settingsFile struct{ Settings }
type config struct {
	Settings
	Folder      string         `json:"folder"`
	Selected    string         `json:"selected"`
	Pages       map[string]int `json:"pages"`
	PageSize    int            `json:"pageSize"`
	BrowseOrder string         `json:"browseOrder"`
	ShuffleSeed uint64         `json:"shuffleSeed"`
	RandomPages map[string]int `json:"randomPages"`
}
type Book struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Error string `json:"error"`
}
type LibraryState struct {
	Folder      string          `json:"folder"`
	Books       []Book          `json:"books"`
	Selected    string          `json:"selected"`
	Words       []wordbook.Word `json:"words"`
	Page        int             `json:"page"`
	Pages       int             `json:"pages"`
	Total       int             `json:"total"`
	Focus       int             `json:"focus"`
	Error       string          `json:"error"`
	BrowseOrder string          `json:"browseOrder"`
	Quiz        QuizState       `json:"quiz"`
}
type Core struct {
	mu           sync.Mutex
	path         string
	cfg          config
	books        []Book
	words        map[string][]wordbook.Word
	focus        int
	scanError    string
	quiz         quizSession
	shuffleCache map[string][]int
	shuffleSeed  uint64
}

func defaultSettings() Settings {
	return Settings{IslandVisible: true, IslandStyle: "capsule", IslandTheme: "dark", IslandHover: true, IslandExpandDelay: 160, IslandCollapseDelay: 420, IslandAnimationDuration: 360, IslandWidth: 520, IslandHeight: 200}
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
	c.quiz = quizSession{}
	c.shuffleCache = nil
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
	page := max(0, min(c.browsePagesLocked()[c.cfg.Selected], max(0, pages-1)))
	start := min(page*pageSize, len(words))
	end := min(start+pageSize, len(words))
	visible := append([]wordbook.Word{}, words[start:end]...)
	if c.cfg.BrowseOrder == "random" {
		indices := c.browseIndicesLocked()
		for i := range visible {
			visible[i] = words[indices[start+i]]
		}
	}
	order := "sequential"
	if c.cfg.BrowseOrder == "random" {
		order = "random"
	}
	return LibraryState{Folder: c.cfg.Folder, Books: c.books, Selected: c.cfg.Selected, Words: visible, Page: page, Pages: pages, Total: len(words), Focus: max(0, min(c.focus, max(0, len(visible)-1))), Error: c.scanError, BrowseOrder: order, Quiz: c.quizSnapshotLocked()}
}

func (c *Core) browsePagesLocked() map[string]int {
	if c.cfg.BrowseOrder == "random" {
		if c.cfg.RandomPages == nil {
			c.cfg.RandomPages = map[string]int{}
		}
		return c.cfg.RandomPages
	}
	if c.cfg.Pages == nil {
		c.cfg.Pages = map[string]int{}
	}
	return c.cfg.Pages
}

// Cache one stable permutation per book. Page changes never reshuffle it.
func (c *Core) browseIndicesLocked() []int {
	if c.shuffleCache == nil || c.shuffleSeed != c.cfg.ShuffleSeed {
		c.shuffleCache = map[string][]int{}
		c.shuffleSeed = c.cfg.ShuffleSeed
	}
	if indices, ok := c.shuffleCache[c.cfg.Selected]; ok {
		return indices
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(c.cfg.Selected))
	rng := rand.New(rand.NewPCG(c.cfg.ShuffleSeed, h.Sum64()))
	indices := rng.Perm(len(c.words[c.cfg.Selected]))
	c.shuffleCache[c.cfg.Selected] = indices
	return indices
}
func (c *Core) snapshot() LibraryState { c.mu.Lock(); defer c.mu.Unlock(); return c.snapshotLocked() }
func (c *Core) publish()               { _ = LibraryEvent.Broadcast(c.snapshot()) }
func (c *Core) change(run func() error) (LibraryState, error) {
	c.mu.Lock()
	// Preferences and navigation are committed together; restore on disk failures.
	old, _ := json.Marshal(c.cfg)
	oldQuiz := c.quiz
	err := run()
	if err == nil {
		err = c.saveLocked()
	}
	if err != nil {
		var restored config
		_ = json.Unmarshal(old, &restored)
		c.cfg = restored
		c.scanLocked()
		c.quiz = oldQuiz
	}
	s := c.snapshotLocked()
	c.mu.Unlock()
	if err == nil {
		_ = LibraryEvent.Broadcast(s)
	}
	return s, err
}
