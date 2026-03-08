// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package giodesktop implements the desktop platform driver using Gio
// (github.com/mlekudev/gio) for window management and GPU rendering.
package giodesktop

import (
	"image"
	"os"
	"runtime"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/events"
	coresystem "cogentcore.org/core/system"
	"cogentcore.org/core/system/composer"
	"cogentcore.org/core/system/driver/base"
	gioapp "github.com/mlekudev/gio/app"
	"github.com/mlekudev/gio/unit"
)

func Init() {
	runtime.LockOSThread()
	TheApp.GetScreens()
	base.Init(TheApp, &TheApp.App)
}

// TheApp is the single [coresystem.App] for the Gio desktop platform.
var TheApp = &App{AppMulti: base.NewAppMulti[*Window]()}

// App is the [coresystem.App] implementation for the Gio desktop platform.
type App struct {
	base.AppMulti[*Window]
}

func (a *App) SendEmptyEvent() {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	for _, w := range a.Windows {
		w.Mu.Lock()
		gw := w.GioWin
		w.Mu.Unlock()
		if gw != nil {
			gw.Invalidate()
		}
	}
}

func (a *App) MainLoop() {
	a.MainQueue = make(chan base.FuncRun)
	a.MainDone = make(chan struct{})
	gioapp.Main()
}

func (a *App) NewWindow(opts *coresystem.NewWindowOptions) (coresystem.Window, error) {
	if opts == nil {
		opts = &coresystem.NewWindowOptions{}
	}
	opts.Fixup()

	w := &Window{
		WindowMulti: base.NewWindowMulti[*App, *composer.ComposerDrawer](a, opts),
	}
	w.This = w

	gioOpts := gioWindowOptions(opts)
	gw := new(gioapp.Window)
	gw.Option(gioOpts...)
	w.GioWin = gw

	w.GioDraw = &GioDrawer{}
	w.Compose = &composer.ComposerDrawer{Drawer: w.GioDraw}

	w.PixelSize = opts.Size
	w.WnSize = opts.Size

	a.Mu.Lock()
	a.Windows = append(a.Windows, w)
	a.Mu.Unlock()

	w.Event.WindowResize()
	w.Event.Window(events.WinShow)
	w.Event.Window(events.ScreenUpdate)
	w.Event.Window(events.WinFocus)

	go w.eventLoop()
	go w.WinLoop()

	return w, nil
}

func (a *App) GetScreens() {
	if len(a.Screens) == 0 {
		a.Screens = []*coresystem.Screen{defaultScreen()}
	}
}

func (a *App) Platform() coresystem.Platforms {
	switch runtime.GOOS {
	case "darwin":
		return coresystem.MacOS
	case "windows":
		return coresystem.Windows
	default:
		return coresystem.Linux
	}
}

func (a *App) DataDir() string {
	dir, err := os.UserConfigDir()
	if errors.Log(err) != nil {
		return ""
	}
	return dir
}

// defaultScreen returns a default screen with reasonable parameters.
func defaultScreen() *coresystem.Screen {
	return &coresystem.Screen{
		PixelSize:        image.Pt(1920, 1080),
		PhysicalSize:     image.Pt(521, 293),
		PhysicalDPI:      96,
		LogicalDPI:       96,
		DevicePixelRatio: 1,
		RefreshRate:      60,
	}
}

// gioWindowOptions converts Cogentcore NewWindowOptions to Gio Options.
func gioWindowOptions(opts *coresystem.NewWindowOptions) []gioapp.Option {
	var gioOpts []gioapp.Option
	if opts.Title != "" {
		gioOpts = append(gioOpts, gioapp.Title(opts.Title))
	}
	if opts.Size.X > 0 && opts.Size.Y > 0 {
		gioOpts = append(gioOpts, gioapp.Size(unit.Dp(opts.Size.X), unit.Dp(opts.Size.Y)))
	}
	if opts.Flags.HasFlag(coresystem.Fullscreen) {
		gioOpts = append(gioOpts, gioapp.Fullscreen.Option())
	}
	if opts.Flags.HasFlag(coresystem.Maximized) {
		gioOpts = append(gioOpts, gioapp.Maximized.Option())
	}
	if opts.Flags.HasFlag(coresystem.Tool) {
		gioOpts = append(gioOpts, gioapp.Decorated(false))
	}
	return gioOpts
}
