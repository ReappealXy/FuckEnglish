package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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
