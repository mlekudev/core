// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js

// Package jsutil provides simple JavaScript interop utilities for WebAssembly builds.
package jsutil

import "syscall/js"

// BytesToJS copies a Go byte slice to a JavaScript Uint8ClampedArray.
func BytesToJS(b []byte) js.Value {
	array := js.Global().Get("Uint8ClampedArray").New(len(b))
	js.CopyBytesToJS(array, b)
	return array
}

// Await waits for a JavaScript Promise to resolve and returns the result.
// The bool return value is true if the promise resolved, false if it rejected.
func Await(promise js.Value) (js.Value, bool) {
	ch := make(chan js.Value, 1)
	errCh := make(chan js.Value, 1)
	promise.Call("then",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			ch <- args[0]
			return nil
		}),
		js.FuncOf(func(this js.Value, args []js.Value) any {
			errCh <- args[0]
			return nil
		}),
	)
	select {
	case result := <-ch:
		return result, true
	case result := <-errCh:
		return result, false
	}
}
