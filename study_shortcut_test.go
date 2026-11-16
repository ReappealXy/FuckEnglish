package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestStudyShortcutValidation(t *testing.T) {
	for _, tc := range []struct {
		input, want string
		key         uintptr
	}{
		{"alt+q", "Alt+Q", 0x51},
		{" shift + ctrl + alt + 7 ", "Ctrl+Alt+Shift+7", 0x37},
		{"win+F8", "Super+F8", 0x77},
		{"super+ctrl+f12", "Ctrl+Super+F12", 0x7B},
		{"control+z", "Ctrl+Z", 0x5A},
	} {
		got, key, err := normalizeStudyShortcut(tc.input)
		if err != nil || got != tc.want || key != tc.key {
			t.Fatalf("normalize %q = %q, %x, %v; want %q, %x", tc.input, got, key, err, tc.want, tc.key)
		}
	}
	for _, invalid := range []string{"", "Q", "Shift+Q", "Alt", "Alt+", "Ctrl+Alt", "Ctrl+Ctrl+Q", "Alt+Alt+Q", "Ctrl+Q+A", "Alt+F0", "Alt+F13", "Alt+F01", "Ctrl+Space", "Alt+字", "Alt+;", "Cmd+Q"} {
		if _, _, err := normalizeStudyShortcut(invalid); err == nil {
			t.Errorf("invalid shortcut %q accepted", invalid)
		}
	}
}

type shortcutRegistryStub struct {
	callbacks map[string]func()
	conflicts map[string]bool
	log       []string
}

func newShortcutRegistrationStub() (*studyShortcutRegistration, *shortcutRegistryStub) {
	s := &shortcutRegistryStub{callbacks: map[string]func(){}, conflicts: map[string]bool{}}
	r := &studyShortcutRegistration{
		register: func(key string, callback func()) error {
			s.log = append(s.log, "register "+key)
			if s.conflicts[key] || s.callbacks[key] != nil {
				return errors.New("already registered")
			}
			s.callbacks[key] = callback
			return nil
		},
		unregister: func(key string) {
			s.log = append(s.log, "unregister "+key)
			delete(s.callbacks, key)
		},
	}
	return r, s
}

func TestStudyShortcutReplacementCommitsOnlyAfterPersistence(t *testing.T) {
	r, native := newShortcutRegistrationStub()
	var pressed []uintptr
	r.onPress = func(key uintptr) { pressed = append(pressed, key) }
	if err := r.replace(defaultStudyShortcut, nil); err != nil {
		t.Fatal(err)
	}
	oldCallback := native.callbacks[defaultStudyShortcut]
	if err := r.replace("Ctrl+Shift+K", func() error {
		if r.shortcut != defaultStudyShortcut || !r.registered || native.callbacks[defaultStudyShortcut] == nil {
			t.Fatal("old binding removed before persistence completed")
		}
		native.callbacks["Ctrl+Shift+K"]()
		if len(pressed) != 0 {
			t.Fatal("candidate binding fired before commit")
		}
		native.log = append(native.log, "persist")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if want := []string{"register Alt+Q", "register Ctrl+Shift+K", "persist", "unregister Alt+Q"}; !reflect.DeepEqual(native.log, want) {
		t.Fatalf("replacement order %v, want %v", native.log, want)
	}
	oldCallback()
	if len(pressed) != 0 {
		t.Fatal("queued old callback fired after commit")
	}
	native.callbacks["Ctrl+Shift+K"]()
	if !reflect.DeepEqual(pressed, []uintptr{0x4B}) || r.shortcut != "Ctrl+Shift+K" || !r.registered {
		t.Fatalf("new binding inactive: presses %v, registration %+v", pressed, r)
	}
	// Saving another setting with the same shortcut does not unregister it.
	native.log = nil
	if err := r.replace("ctrl+shift+k", func() error { native.log = append(native.log, "persist"); return nil }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(native.log, []string{"persist"}) {
		t.Fatalf("unchanged binding was re-registered: %v", native.log)
	}
}

func TestStudyShortcutConflictAndWriteFailurePreserveWorkingPreference(t *testing.T) {
	r, native := newShortcutRegistrationStub()
	c := testCore(t)
	if err := c.saveLocked(); err != nil {
		t.Fatal(err)
	}
	if err := r.replace(defaultStudyShortcut, nil); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(c.path)
	if err != nil {
		t.Fatal(err)
	}
	persistCalls := 0
	persist := func() error {
		persistCalls++
		_, err := c.change(func() error { c.cfg.StudyShortcut = "Ctrl+Alt+L"; return nil })
		return err
	}
	native.conflicts["Ctrl+Alt+L"] = true
	if err = r.replace("Ctrl+Alt+L", persist); err == nil || !strings.Contains(err.Error(), "占用") {
		t.Fatalf("expected readable conflict error, got %v", err)
	}
	if persistCalls != 0 {
		t.Fatal("conflicted shortcut was written to disk")
	}
	delete(native.conflicts, "Ctrl+Alt+L")
	goodPath := c.path
	c.path = filepath.Join(t.TempDir(), "missing", "settings.json")
	if err = r.replace("Ctrl+Alt+L", persist); err == nil {
		t.Fatal("expected write failure")
	}
	if persistCalls != 1 || r.shortcut != defaultStudyShortcut || !r.registered || native.callbacks[defaultStudyShortcut] == nil || native.callbacks["Ctrl+Alt+L"] != nil || c.cfg.StudyShortcut != defaultStudyShortcut {
		t.Fatalf("failed replacement changed active/persisted settings: calls=%d, registration=%+v, config=%q", persistCalls, r, c.cfg.StudyShortcut)
	}
	after, err := os.ReadFile(goodPath)
	if err != nil || string(before) != string(after) {
		t.Fatalf("original settings changed on failure: %v", err)
	}
	// A shortcut that failed on startup can be retried once its conflict clears.
	r, native = newShortcutRegistrationStub()
	r.shortcut, r.error = defaultStudyShortcut, "occupied on startup"
	if err = r.replace(defaultStudyShortcut, nil); err != nil || !r.registered || r.error != "" {
		t.Fatalf("startup registration did not recover: %+v, %v", r, err)
	}
}

func TestStudyShortcutMigrationAndRestart(t *testing.T) {
	for _, tc := range []struct {
		input any
		want  string
	}{
		{nil, defaultStudyShortcut}, // Old versions did not write this key.
		{"", defaultStudyShortcut},
		{"Q", defaultStudyShortcut},
		{"alt+shift+k", "Alt+Shift+K"},
	} {
		c := testCore(t)
		legacy := map[string]any{"folder": c.cfg.Folder, "selected": "list2.xlsx", "pages": map[string]int{"list2.xlsx": 4}, "pageSize": 5, "islandTheme": "light"}
		if tc.input != nil {
			legacy["studyShortcut"] = tc.input
		}
		data, _ := json.Marshal(legacy)
		if err := os.WriteFile(c.path, data, 0600); err != nil {
			t.Fatal(err)
		}
		fresh := &Core{}
		if err := fresh.Init(filepath.Dir(c.path)); err != nil {
			t.Fatal(err)
		}
		if fresh.cfg.StudyShortcut != tc.want || fresh.cfg.Selected != "list2.xlsx" || fresh.cfg.Pages["list2.xlsx"] != 4 || fresh.cfg.IslandTheme != "light" {
			t.Fatalf("migration changed preferences/navigation: %+v", fresh.cfg)
		}
		restarted := &Core{}
		if err := restarted.Init(filepath.Dir(c.path)); err != nil {
			t.Fatal(err)
		}
		if restarted.cfg.StudyShortcut != tc.want {
			t.Fatalf("shortcut was not saved through restart: %q, want %q", restarted.cfg.StudyShortcut, tc.want)
		}
	}
}
