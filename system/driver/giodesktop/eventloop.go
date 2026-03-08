// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giodesktop

import (
	"image"

	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	gioapp "github.com/mlekudev/gio/app"
	"github.com/mlekudev/gio/io/key"
	"github.com/mlekudev/gio/io/pointer"
)

// eventLoop runs in its own goroutine, reading events from the Gio window
// and translating them into Cogentcore events pushed to the window's event deque.
func (w *Window) eventLoop() {
	for {
		// Snapshot the Gio window pointer under a brief lock to avoid
		// racing with Close() which nils GioWin.
		w.Mu.Lock()
		gw := w.GioWin
		w.Mu.Unlock()
		if gw == nil {
			return
		}
		ev := gw.Event()
		switch e := ev.(type) {
		case gioapp.DestroyEvent:
			w.handleDestroy(e)
			return
		case gioapp.FrameEvent:
			w.handleFrame(e)
		case gioapp.ConfigEvent:
			w.handleConfig(e)
		case key.Event:
			w.handleKey(e)
		case key.EditEvent:
			w.handleEdit(e)
		case pointer.Event:
			w.handlePointer(e)
		}
	}
}

func (w *Window) handleDestroy(e gioapp.DestroyEvent) {
	w.Event.Window(events.WinClose)
}

func (w *Window) handleFrame(e gioapp.FrameEvent) {
	w.updateGeometryFromGio(e.Size, e.Metric)
	// Set the frame function so the drawer can submit ops when composition completes.
	w.GioDraw.SetFrameFunc(e.Frame)
}

func (w *Window) handleConfig(e gioapp.ConfigEvent) {
	changed := false
	w.Mu.Lock()
	if e.Config.Size != w.PixelSize {
		w.PixelSize = e.Config.Size
		w.WnSize = e.Config.Size
		changed = true
	}
	w.Mu.Unlock()
	if changed {
		w.Event.WindowResize()
	}
	if e.Config.Focused {
		w.Event.Window(events.WinFocus)
	}
}

func (w *Window) handleKey(e key.Event) {
	code := gioKeyCode(e.Name)
	rn := gioKeyRune(e.Name)
	mods := gioMods(e.Modifiers)

	var typ events.Types
	switch e.State {
	case key.Press:
		typ = events.KeyDown
	case key.Release:
		typ = events.KeyUp
	default:
		return
	}
	w.Event.Key(typ, rn, code, mods)
}

func (w *Window) handleEdit(e key.EditEvent) {
	for _, r := range e.Text {
		w.Event.KeyChord(r, 0, 0)
	}
}

func (w *Window) handlePointer(e pointer.Event) {
	pos := image.Pt(int(e.Position.X), int(e.Position.Y))
	mods := gioMods(e.Modifiers)

	switch e.Kind {
	case pointer.Press:
		but := gioButton(e.Buttons)
		w.Event.MouseButton(events.MouseDown, but, pos, mods)
	case pointer.Release:
		but := gioButton(e.Buttons)
		w.Event.MouseButton(events.MouseUp, but, pos, mods)
	case pointer.Move, pointer.Drag:
		w.Event.MouseMove(pos)
	case pointer.Scroll:
		w.Event.Scroll(pos, math32.Vec2(e.Scroll.X, e.Scroll.Y))
	}
}
