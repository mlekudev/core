// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giodesktop

import (
	"unicode"

	"cogentcore.org/core/events"
	corekey "cogentcore.org/core/events/key"
	giokey "github.com/mlekudev/gio/io/key"
	"github.com/mlekudev/gio/io/pointer"
)

// gioKeyCode maps a Gio key name to a Cogentcore key code.
func gioKeyCode(name giokey.Name) corekey.Codes {
	if len(name) == 1 {
		r := rune(name[0])
		r = unicode.ToUpper(r)
		switch {
		case r >= 'A' && r <= 'Z':
			return corekey.CodeA + corekey.Codes(r-'A')
		case r >= '0' && r <= '9':
			if r == '0' {
				return corekey.Code0
			}
			return corekey.Code1 + corekey.Codes(r-'1')
		}
	}
	switch name {
	case giokey.NameLeftArrow:
		return corekey.CodeLeftArrow
	case giokey.NameRightArrow:
		return corekey.CodeRightArrow
	case giokey.NameUpArrow:
		return corekey.CodeUpArrow
	case giokey.NameDownArrow:
		return corekey.CodeDownArrow
	case giokey.NameReturn, giokey.NameEnter:
		return corekey.CodeReturnEnter
	case giokey.NameEscape:
		return corekey.CodeEscape
	case giokey.NameHome:
		return corekey.CodeHome
	case giokey.NameEnd:
		return corekey.CodeEnd
	case giokey.NameDeleteBackward:
		return corekey.CodeBackspace
	case giokey.NameDeleteForward:
		return corekey.CodeDelete
	case giokey.NamePageUp:
		return corekey.CodePageUp
	case giokey.NamePageDown:
		return corekey.CodePageDown
	case giokey.NameTab:
		return corekey.CodeTab
	case giokey.NameSpace:
		return corekey.CodeSpacebar
	case giokey.NameF1:
		return corekey.CodeF1
	case giokey.NameF2:
		return corekey.CodeF2
	case giokey.NameF3:
		return corekey.CodeF3
	case giokey.NameF4:
		return corekey.CodeF4
	case giokey.NameF5:
		return corekey.CodeF5
	case giokey.NameF6:
		return corekey.CodeF6
	case giokey.NameF7:
		return corekey.CodeF7
	case giokey.NameF8:
		return corekey.CodeF8
	case giokey.NameF9:
		return corekey.CodeF9
	case giokey.NameF10:
		return corekey.CodeF10
	case giokey.NameF11:
		return corekey.CodeF11
	case giokey.NameF12:
		return corekey.CodeF12
	case giokey.NameCtrl:
		return corekey.CodeLeftControl
	case giokey.NameShift:
		return corekey.CodeLeftShift
	case giokey.NameAlt:
		return corekey.CodeLeftAlt
	case giokey.NameSuper, giokey.NameCommand:
		return corekey.CodeLeftMeta
	}
	return corekey.CodeUnknown
}

// gioKeyRune returns the rune associated with a Gio key name,
// taking shift state into account for single-character keys.
func gioKeyRune(name giokey.Name) rune {
	if len(name) == 1 {
		return rune(name[0])
	}
	switch name {
	case giokey.NameReturn, giokey.NameEnter:
		return '\n'
	case giokey.NameTab:
		return '\t'
	case giokey.NameSpace:
		return ' '
	case giokey.NameDeleteBackward:
		return '\b'
	case giokey.NameEscape:
		return 0x1B
	}
	return 0
}

// gioMods maps Gio modifier flags to Cogentcore modifier flags.
func gioMods(m giokey.Modifiers) corekey.Modifiers {
	var mods corekey.Modifiers
	if m.Contain(giokey.ModCtrl) {
		mods |= 1 << corekey.Control
	}
	if m.Contain(giokey.ModCommand) {
		mods |= 1 << corekey.Meta
	}
	if m.Contain(giokey.ModShift) {
		mods |= 1 << corekey.Shift
	}
	if m.Contain(giokey.ModAlt) {
		mods |= 1 << corekey.Alt
	}
	if m.Contain(giokey.ModSuper) {
		mods |= 1 << corekey.Meta
	}
	return mods
}

// gioButton maps a Gio pointer button to a Cogentcore mouse button.
func gioButton(btns pointer.Buttons) events.Buttons {
	switch {
	case btns.Contain(pointer.ButtonPrimary):
		return events.Left
	case btns.Contain(pointer.ButtonSecondary):
		return events.Right
	case btns.Contain(pointer.ButtonTertiary):
		return events.Middle
	}
	return events.NoButton
}
