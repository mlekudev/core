// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giodesktop

import (
	"image"
	"math"
	"os"

	"cogentcore.org/core/events"
	corekey "cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	gioapp "github.com/mlekudev/gio/app"
	"github.com/mlekudev/gio/io/key"
	"github.com/mlekudev/gio/io/pointer"
)

// eventLoop runs in its own goroutine, reading events from the Gio window
// and translating them into Cogentcore events pushed to the window's event deque.
//
// In Gio, pointer and key events are NOT delivered through Window.Event().
// They flow through the input router and are read via FrameEvent.Source.Event()
// using filters that match the input area registered in the op stream.
func (w *Window) eventLoop() {
	for {
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
		}
	}
}

func (w *Window) handleDestroy(e gioapp.DestroyEvent) {
	w.Event.Window(events.WinClose)
	// Signal WinLoop to stop and exit the process.
	select {
	case w.WinClose <- struct{}{}:
	default:
	}
	os.Exit(0)
}

func (w *Window) handleFrame(e gioapp.FrameEvent) {
	w.updateGeometryFromGio(e.Size, e.Metric)

	// Read input events from Gio's router. These were collected since
	// the last frame based on the input area we registered in the ops.
	w.processInputEvents(e)

	// Trigger a paint event so the compositor runs.
	w.Event.WindowPaint()

	// Block until composition finishes and e.Frame is called.
	w.GioDraw.FrameReady(e.Frame, e.Size)
}

// processInputEvents reads pointer and key events from the Gio input
// source and translates them into Cogentcore events.
func (w *Window) processInputEvents(e gioapp.FrameEvent) {
	for {
		ev, ok := e.Source.Event(
			pointer.Filter{
				Target:  &inputTag,
				Kinds:   pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Scroll | pointer.Enter | pointer.Leave | pointer.Cancel,
				ScrollX: pointer.ScrollRange{Min: -math.MaxInt32, Max: math.MaxInt32},
				ScrollY: pointer.ScrollRange{Min: -math.MaxInt32, Max: math.MaxInt32},
			},
			key.Filter{
				Focus:    nil, // receive regardless of focus
				Optional: key.ModCtrl | key.ModCommand | key.ModShift | key.ModAlt | key.ModSuper,
			},
		)
		if !ok {
			break
		}
		switch pe := ev.(type) {
		case pointer.Event:
			w.handlePointer(pe)
		case key.Event:
			w.handleKey(pe)
		case key.EditEvent:
			w.handleEdit(pe)
		}
	}
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
	} else {
		w.Event.Window(events.WinFocusLost)
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
		w.Event.Key(typ, rn, code, mods)
		// Also send KeyChord for printable keys on press.
		if rn > 0 && !mods.HasFlag(corekey.Control) && !mods.HasFlag(corekey.Meta) {
			w.Event.KeyChord(rn, code, mods)
		}
	case key.Release:
		typ = events.KeyUp
		w.Event.Key(typ, rn, code, mods)
	}
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
	case pointer.Enter:
		w.Event.MouseMove(pos)
	case pointer.Leave:
		// Could send a leave event if needed.
	}
}
