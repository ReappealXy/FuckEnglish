package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/egoist/mygo"
)

const defaultStudyShortcut = "Alt+Q"

type StudyShortcutStatus struct {
	Shortcut   string `json:"shortcut"`
	Registered bool   `json:"registered"`
	Error      string `json:"error"`
}

// The registration and window state are owned by the MyGo main thread.
// Keeping the old binding until a successful disk write makes failed changes
// leave both the saved preference and its working shortcut intact.
type studyShortcutRegistration struct {
	shortcut   string
	registered bool
	error      string
	register   func(string, func()) error
	unregister func(string)
	onPress    func(uintptr)
}

var globalStudyShortcut = studyShortcutRegistration{
	register: mygo.GlobalShortcut.Register, unregister: mygo.GlobalShortcut.Unregister,
}

// Use a small, predictable subset of native accelerators. A modifier is
// required so ordinary typing in another app never becomes a view switch.
func normalizeStudyShortcut(value string) (string, uintptr, error) {
	parts := strings.Split(strings.TrimSpace(value), "+")
	if len(parts) < 2 {
		return "", 0, fmt.Errorf("切换快捷键需要 Ctrl、Alt 或 Win，加一个字母、数字或 F1～F12")
	}
	modifiers := map[string]bool{}
	for _, part := range parts[:len(parts)-1] {
		var modifier string
		switch strings.ToLower(strings.TrimSpace(part)) {
		case "ctrl", "control":
			modifier = "Ctrl"
		case "alt":
			modifier = "Alt"
		case "shift":
			modifier = "Shift"
		case "super", "win", "meta":
			modifier = "Super"
		default:
			return "", 0, fmt.Errorf("不支持的快捷键修饰键：%s", strings.TrimSpace(part))
		}
		if modifiers[modifier] {
			return "", 0, fmt.Errorf("快捷键修饰键不能重复：%s", modifier)
		}
		modifiers[modifier] = true
	}
	if !modifiers["Ctrl"] && !modifiers["Alt"] && !modifiers["Super"] {
		return "", 0, fmt.Errorf("切换快捷键至少需要包含 Ctrl、Alt 或 Win，避免影响正常打字")
	}
	key := strings.ToUpper(strings.TrimSpace(parts[len(parts)-1]))
	var virtualKey uintptr
	if len(key) == 1 && ((key[0] >= 'A' && key[0] <= 'Z') || (key[0] >= '0' && key[0] <= '9')) {
		virtualKey = uintptr(key[0])
	} else if strings.HasPrefix(key, "F") {
		number, err := strconv.Atoi(key[1:])
		if err == nil && number >= 1 && number <= 12 && key == fmt.Sprintf("F%d", number) {
			virtualKey = uintptr(0x70 + number - 1)
		}
	}
	if virtualKey == 0 {
		return "", 0, fmt.Errorf("快捷键主键请选择字母 A～Z、数字 0～9 或 F1～F12")
	}
	canonical := []string{}
	for _, modifier := range []string{"Ctrl", "Alt", "Shift", "Super"} {
		if modifiers[modifier] {
			canonical = append(canonical, modifier)
		}
	}
	return strings.Join(append(canonical, key), "+"), virtualKey, nil
}

func (r *studyShortcutRegistration) replace(value string, persist func() error) error {
	shortcut, key, err := normalizeStudyShortcut(value)
	if err != nil {
		return err
	}
	if r.registered && r.shortcut == shortcut {
		if persist != nil {
			if err = persist(); err != nil {
				return err
			}
		}
		r.error = ""
		return nil
	}
	if err = r.register(shortcut, func() {
		// A newly registered candidate is inactive until persistence succeeds;
		// queued callbacks from an old registration are ignored after a switch.
		if r.registered && r.shortcut == shortcut && r.onPress != nil {
			r.onPress(key)
		}
	}); err != nil {
		return fmt.Errorf("无法启用快捷键 %s，可能已被其他程序或系统占用，请换一组组合键：%w", shortcut, err)
	}
	if persist != nil {
		if err = persist(); err != nil {
			r.unregister(shortcut)
			return err
		}
	}
	old, hadOld := r.shortcut, r.registered
	r.shortcut, r.registered, r.error = shortcut, true, ""
	if hadOld {
		r.unregister(old)
	}
	return nil
}

func setupStudyShortcut(core *Core) {
	mygo.RunOnMain(func() {
		globalStudyShortcut.onPress = func(key uintptr) {
			if applicationQuitting || !acceptStudyTogglePress(key) {
				return
			}
			if studyActive {
				returnToCapsule(core)
			} else {
				showStudyWindow(core)
			}
		}
		shortcut := core.settings().StudyShortcut
		if err := globalStudyShortcut.replace(shortcut, nil); err != nil {
			globalStudyShortcut.shortcut = shortcut
			globalStudyShortcut.error = err.Error()
			log.Println("study shortcut:", err)
		}
	})
}

func stopStudyShortcut() {
	mygo.RunOnMain(func() {
		if globalStudyShortcut.registered {
			globalStudyShortcut.unregister(globalStudyShortcut.shortcut)
			globalStudyShortcut.registered = false
		}
	})
}

func (p *Preferences) ShortcutStatus() StudyShortcutStatus {
	var result StudyShortcutStatus
	mygo.RunOnMain(func() {
		result = StudyShortcutStatus{Shortcut: globalStudyShortcut.shortcut, Registered: globalStudyShortcut.registered, Error: globalStudyShortcut.error}
	})
	return result
}
