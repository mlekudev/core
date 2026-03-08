// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giodesktop

import (
	"sync"

	"cogentcore.org/core/base/fileinfo/mimedata"
)

// Clipboard implements system.Clipboard for the Gio desktop platform.
// Gio handles clipboard through its event/op system, so this is a
// minimal implementation that stores clipboard data locally and
// defers to system clipboard through Gio ops when available.
type Clipboard struct {
	mu   sync.Mutex
	data mimedata.Mimes
}

func (c *Clipboard) IsEmpty() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.data) == 0
}

func (c *Clipboard) Read(types []string) mimedata.Mimes {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data
}

func (c *Clipboard) Write(data mimedata.Mimes) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	return nil
}

func (c *Clipboard) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = nil
}
