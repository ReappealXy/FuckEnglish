//go:build !windows

package main

import "github.com/egoist/mygo"

func configureStudyFrame(*mygo.Window) error   { return nil }
func configureStudyTaskbar(*mygo.Window) error { return nil }
