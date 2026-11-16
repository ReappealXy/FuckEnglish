package main

import (
	"log"

	"github.com/egoist/mygo"
)

// Desired page state is retained while the capsule is hidden, so returning to
// the capsule does not depend on a new frontend render to restore the shortcut.
var islandStudyAllowed bool
var islandStudyRegistered bool

func (i *Island) SetStudyShortcut(enabled bool) {
	mygo.RunOnMain(func() {
		islandStudyAllowed = enabled
		updateIslandStudyShortcut(i.core)
	})
}

func (i *Island) ToggleStudyShortcut(fromStudy bool) {
	mygo.RunOnMain(func() { toggleStudyFromShortcut(i.core, fromStudy) })
}

func toggleStudyFromShortcut(core *Core, fromStudy bool) {
	// Stale page events and a held key must not reverse the transition.
	if studyActive != fromStudy || applicationQuitting {
		return
	}
	if !fromStudy && (!islandPointerInside || !islandStudyAllowed) {
		return
	}
	if !acceptStudyTogglePress() {
		return
	}
	if fromStudy {
		returnToCapsule(core)
	} else {
		showStudyWindow(core)
	}
}

func updateIslandStudyShortcut(core *Core) {
	active := !studyActive && islandReady && islandPointerInside && islandStudyAllowed && islandWin != nil && !islandWin.IsDestroyed() && islandWin.IsVisible()
	if !active {
		if islandStudyRegistered {
			mygo.GlobalShortcut.Unregister("Q")
			islandStudyRegistered = false
		}
		return
	}
	if islandStudyRegistered || mygo.GlobalShortcut.IsRegistered("Q") {
		return
	}
	if err := mygo.GlobalShortcut.Register("Q", func() {
		if !studyActive && islandPointerInside && islandStudyAllowed {
			toggleStudyFromShortcut(core, false)
		}
	}); err != nil {
		log.Println("island study shortcut:", err)
		return
	}
	islandStudyRegistered = true
}
