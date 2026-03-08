// Copyright 2018 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// based on golang.org/x/exp/shiny:
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(android || ios || js)

package driver

import (
	"os"
	"slices"
	"testing"

	"cogentcore.org/core/system/driver/giodesktop"
	"cogentcore.org/core/system/driver/offscreen"
)

func init() {
	if testing.Testing() || slices.Contains(os.Args, "-nogui") {
		offscreen.Init()
		return
	}
	giodesktop.Init()
}
