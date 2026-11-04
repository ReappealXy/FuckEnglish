//go:build !windows

package main

import "github.com/egoist/mygo"

func configureIslandOverlay(*mygo.Window) error     { return nil }
func startIslandPointer(*mygo.Window, *Core) func() { return func() {} }
