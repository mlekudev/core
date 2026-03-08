// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giodesktop

import (
	"image"

	coresystem "cogentcore.org/core/system"
	"cogentcore.org/core/system/composer"
	"cogentcore.org/core/system/driver/base"
	gioapp "github.com/mlekudev/gio/app"
	giosystem "github.com/mlekudev/gio/io/system"
	"github.com/mlekudev/gio/unit"
)

// Window is the implementation of [coresystem.Window] for the Gio desktop platform.
type Window struct {
	base.WindowMulti[*App, *composer.ComposerDrawer]

	// GioWin is the underlying Gio window.
	GioWin *gioapp.Window

	// GioDraw is the Gio-backed drawer for compositing.
	GioDraw *GioDrawer
}

func (w *Window) SetTitle(title string) {
	if w.IsClosed() {
		return
	}
	w.Titl = title
	w.GioWin.Option(gioapp.Title(title))
}

func (w *Window) SetWinSize(sz image.Point) {
	if w.IsClosed() {
		return
	}
	w.WindowMulti.SetWinSize(sz)
	w.GioWin.Option(gioapp.Size(unit.Dp(sz.X), unit.Dp(sz.Y)))
}

func (w *Window) Raise() {
	if w.IsClosed() {
		return
	}
	w.GioWin.Perform(giosystem.ActionRaise)
}

func (w *Window) Minimize() {
	if w.IsClosed() {
		return
	}
	w.GioWin.Option(gioapp.Minimized.Option())
}

func (w *Window) Close() {
	if w == nil {
		return
	}
	w.Mu.Lock()
	gw := w.GioWin
	w.GioWin = nil
	w.Mu.Unlock()
	w.Window.Close()
	if gw != nil {
		gw.Perform(giosystem.ActionClose)
	}
}

func (w *Window) SendPaintEvent() {
	w.This.Events().WindowPaint()
	if w.GioWin != nil {
		w.GioWin.Invalidate()
	}
}

func (w *Window) Screen() *coresystem.Screen {
	if len(TheApp.Screens) == 0 {
		return nil
	}
	sc := TheApp.Screens[0]
	w.Mu.Lock()
	w.PhysDPI = sc.PhysicalDPI
	if w.LogDPI == 0 {
		w.LogDPI = sc.LogicalDPI
	}
	w.Mu.Unlock()
	return sc
}

func (w *Window) IsClosed() bool {
	return w == nil || w.This == nil || w.GioWin == nil
}

// updateGeometryFromGio updates window geometry from Gio frame event data.
func (w *Window) updateGeometryFromGio(size image.Point, metric unit.Metric) {
	w.Mu.Lock()
	changed := w.PixelSize != size
	w.PixelSize = size
	w.WnSize = size
	dpi := metric.PxPerDp * 96
	w.PhysDPI = dpi
	if w.LogDPI == 0 {
		w.LogDPI = dpi
	}
	w.DevicePixelRatio = metric.PxPerDp
	w.Mu.Unlock()
	if changed {
		w.Event.WindowResize()
	}
}
