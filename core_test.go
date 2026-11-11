package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func testCore(t *testing.T) *Core {
	t.Helper()
	folder, err := filepath.Abs("words")
	if err != nil {
		t.Fatal(err)
	}
	c := &Core{path: filepath.Join(t.TempDir(), "settings.json"), cfg: config{Settings: defaultSettings(), Folder: folder, Selected: "list1.xlsx", Pages: map[string]int{}}}
	c.scanLocked()
	return c
}
func TestPaginationAndBookResume(t *testing.T) {
	c := testCore(t)
	l := &Library{c}
	s := l.Snapshot()
	if s.Total != 50 || s.Pages != 10 || len(s.Words) != 5 || s.Words[0].English != "pioneer" {
		t.Fatalf("first page: %+v", s)
	}
	s, e := l.Page(999)
	if e != nil {
		t.Fatal(e)
	}
	if s.Page != 9 || len(s.Words) != 5 || s.Words[4].English != "rectify" {
		t.Fatalf("last page: %+v", s)
	}
	if _, e = l.Select("list2.xlsx"); e != nil {
		t.Fatal(e)
	}
	s, e = l.Select("list1.xlsx")
	if e != nil || s.Page != 9 {
		t.Fatalf("resume: %+v %v", s, e)
	}
	var disk config
	b, _ := os.ReadFile(c.path)
	if e = json.Unmarshal(b, &disk); e != nil || disk.Pages["list1.xlsx"] != 9 {
		t.Fatalf("disk: %+v %v", disk, e)
	}
	s, e = l.Page(-3)
	if e != nil || s.Page != 0 {
		t.Fatal("negative page not clamped")
	}
	if _, e = l.Select("../list1.xlsx"); e == nil {
		t.Fatal("invalid name accepted")
	}
}
func TestEmptyInvalidAndMissingFolder(t *testing.T) {
	c := testCore(t)
	l := &Library{c}
	empty := t.TempDir()
	s, e := l.SetFolder(empty)
	if e != nil || s.Total != 0 || s.Pages != 0 || len(s.Words) != 0 {
		t.Fatalf("empty: %+v %v", s, e)
	}
	os.WriteFile(filepath.Join(empty, "bad.xlsx"), []byte("broken"), 0600)
	s, e = l.Refresh()
	if e != nil || len(s.Books) != 1 || s.Books[0].Error == "" {
		t.Fatalf("invalid: %+v %v", s, e)
	}
	if _, e = l.SetFolder(filepath.Join(empty, "missing")); e == nil {
		t.Fatal("missing folder accepted")
	}
	if l.Snapshot().Folder != empty {
		t.Fatal("bad folder replaced valid selection")
	}
}
func TestSaveFailureRollsBackSelection(t *testing.T) {
	c := testCore(t)
	c.path = filepath.Join(t.TempDir(), "missing", "settings.json")
	l := &Library{c}
	before := l.Snapshot()
	if _, e := l.Select("list2.xlsx"); e == nil {
		t.Fatal("save should fail")
	}
	if l.Snapshot().Selected != before.Selected {
		t.Fatal("failed save changed selected book")
	}
}

func TestNaturalBookNameOrder(t *testing.T) {
	names := []string{
		"list10.xlsx",
		"list00000000000000000000000000000000000000000000000001.xlsx",
		"list2.xlsx",
		"LIST1.xlsx",
		"list9.xlsx",
		"list01.xlsx",
		"list1.xlsx",
	}
	sort.Slice(names, func(i, j int) bool { return naturalNameLess(names[i], names[j]) })
	want := []string{
		"LIST1.xlsx",
		"list1.xlsx",
		"list01.xlsx",
		"list00000000000000000000000000000000000000000000000001.xlsx",
		"list2.xlsx",
		"list9.xlsx",
		"list10.xlsx",
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("natural order = %v, want %v", names, want)
		}
	}
	for _, pair := range [][2]string{
		{"list99999999999999999999999999999999999999.xlsx", "list100000000000000000000000000000000000000.xlsx"},
		{"list2-part9.xlsx", "list2-part10.xlsx"},
		{"词表2.xlsx", "词表10.xlsx"},
		{"list0.xlsx", "list00.xlsx"},
		{"list.xlsx", "list2.xlsx"},
	} {
		if !naturalNameLess(pair[0], pair[1]) || naturalNameLess(pair[1], pair[0]) {
			t.Errorf("expected %q before %q", pair[0], pair[1])
		}
	}
}

func TestScanBookNaturalOrder(t *testing.T) {
	folder := t.TempDir()
	data, err := examples.ReadFile("words/list1.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"list10.xlsx", "list2.xlsx", "list1.xlsx", "list9.xlsx"} {
		if err := os.WriteFile(filepath.Join(folder, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	c := &Core{cfg: config{Folder: folder, Selected: "list9.xlsx", Pages: map[string]int{"list9.xlsx": 3}}}
	c.scanLocked()
	want := []string{"list1.xlsx", "list2.xlsx", "list9.xlsx", "list10.xlsx"}
	if len(c.books) != len(want) {
		t.Fatalf("found %d books, want %d", len(c.books), len(want))
	}
	for i := range want {
		if c.books[i].Name != want[i] {
			t.Fatalf("scanned book order = %v, want %v", c.books, want)
		}
	}
	state := c.snapshotLocked()
	if state.Selected != "list9.xlsx" || state.Page != 3 || len(state.Words) != pageSize {
		t.Fatalf("sorting changed selection or resume page: %+v", state)
	}
}

func TestOldSettingsKeepOpaqueDefaults(t *testing.T) {
	dir := t.TempDir()
	folder, err := filepath.Abs("words")
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]any{"folder": folder, "selected": "list2.xlsx", "pages": map[string]int{"list2.xlsx": 4}, "pageSize": 5, "islandTheme": "light"}
	data, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	c := &Core{}
	if err := c.Init(dir); err != nil {
		t.Fatal(err)
	}
	if c.cfg.BackgroundTransparency != 0 || c.cfg.TextTransparency != 0 {
		t.Fatalf("legacy settings became transparent: %+v", c.cfg.Settings)
	}
	if s := c.snapshot(); s.Folder != folder || s.Selected != "list2.xlsx" || s.Page != 4 || c.cfg.IslandTheme != "light" {
		t.Fatalf("legacy navigation or theme changed: %+v, %q", s, c.cfg.IslandTheme)
	}
}

func TestTransparencyBoundsPersistenceAndRollback(t *testing.T) {
	for _, values := range [][4]int{{-8, 130, 0, 100}, {130, -8, 100, 0}, {60, 35, 60, 35}} {
		c := testCore(t)
		s := c.cfg.Settings
		s.BackgroundTransparency, s.TextTransparency = values[0], values[1]
		s = normalizeTransparency(s)
		if _, err := c.change(func() error { c.cfg.Settings = s; return nil }); err != nil {
			t.Fatal(err)
		}
		fresh := &Core{}
		if err := fresh.Init(filepath.Dir(c.path)); err != nil {
			t.Fatal(err)
		}
		if fresh.cfg.BackgroundTransparency != values[2] || fresh.cfg.TextTransparency != values[3] {
			t.Fatalf("transparency was not clamped and restored: %+v", fresh.cfg.Settings)
		}
		fresh.path = filepath.Join(t.TempDir(), "missing", "settings.json")
		before := fresh.cfg.Settings
		if _, err := fresh.change(func() error { fresh.cfg.BackgroundTransparency = 12; fresh.cfg.TextTransparency = 89; return nil }); err == nil {
			t.Fatal("expected save failure")
		}
		if fresh.cfg.Settings != before {
			t.Fatalf("failed save changed transparency: %+v, want %+v", fresh.cfg.Settings, before)
		}
	}
	// A hand-edited config is clamped on read as well as through Preferences.Save.
	c := testCore(t)
	c.cfg.BackgroundTransparency, c.cfg.TextTransparency = -10, 200
	if err := c.saveLocked(); err != nil {
		t.Fatal(err)
	}
	fresh := &Core{}
	if err := fresh.Init(filepath.Dir(c.path)); err != nil {
		t.Fatal(err)
	}
	if fresh.cfg.BackgroundTransparency != 0 || fresh.cfg.TextTransparency != 100 {
		t.Fatalf("out-of-range disk settings were not clamped: %+v", fresh.cfg.Settings)
	}
}
