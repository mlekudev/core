// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giodesktop

import (
	"cogentcore.org/core/cursors"
	"github.com/mlekudev/gio/io/pointer"
)

// gioFromCoreCursor maps a Cogentcore cursor enum to a Gio pointer.Cursor.
func gioFromCoreCursor(c cursors.Cursor) pointer.Cursor {
	switch c {
	case cursors.Arrow:
		return pointer.CursorDefault
	case cursors.Text:
		return pointer.CursorText
	case cursors.Pointer:
		return pointer.CursorPointer
	case cursors.Crosshair:
		return pointer.CursorCrosshair
	case cursors.Move:
		return pointer.CursorAllScroll
	case cursors.ResizeN:
		return pointer.CursorNorthResize
	case cursors.ResizeS:
		return pointer.CursorSouthResize
	case cursors.ResizeE:
		return pointer.CursorEastResize
	case cursors.ResizeW:
		return pointer.CursorWestResize
	case cursors.ResizeNE:
		return pointer.CursorNorthEastResize
	case cursors.ResizeNW:
		return pointer.CursorNorthWestResize
	case cursors.ResizeSE:
		return pointer.CursorSouthEastResize
	case cursors.ResizeSW:
		return pointer.CursorSouthWestResize
	case cursors.ResizeEW:
		return pointer.CursorEastWestResize
	case cursors.ResizeNS:
		return pointer.CursorNorthSouthResize
	case cursors.ResizeNESW:
		return pointer.CursorNorthEastSouthWestResize
	case cursors.ResizeNWSE:
		return pointer.CursorNorthWestSouthEastResize
	case cursors.NotAllowed:
		return pointer.CursorNotAllowed
	case cursors.Wait:
		return pointer.CursorWait
	case cursors.Progress:
		return pointer.CursorProgress
	case cursors.Grab:
		return pointer.CursorGrab
	case cursors.Grabbing:
		return pointer.CursorGrabbing
	default:
		return pointer.CursorDefault
	}
}
