// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giodesktop

import (
	"image"
	"image/draw"
	"sync"
	"time"

	"cogentcore.org/core/math32"
	"github.com/mlekudev/gio/f32"
	"github.com/mlekudev/gio/op"
	"github.com/mlekudev/gio/op/clip"
	"github.com/mlekudev/gio/op/paint"
)

// GioDrawer implements composer.Drawer using Gio's operation stream
// to composite pre-rendered images onto the window surface.
//
// Synchronization model: Gio requires that e.Frame(&ops) is called
// before the next Event() call. The event loop goroutine calls
// FrameReady() when a FrameEvent arrives, which blocks until
// the composition goroutine finishes (End calls e.Frame via submit).
type GioDrawer struct {
	mu sync.Mutex

	// ops is the Gio operation list built during Start/Copy/End.
	ops op.Ops

	// ready indicates that ops have been built and are ready for frame submission.
	ready bool

	// frameFunc is set by the event loop's FrameEvent handler;
	// calling it submits the ops to Gio's GPU pipeline.
	frameFunc func(*op.Ops)

	// frameDone is signaled when a frame has been submitted (or skipped).
	// The event loop blocks on this after setting the frameFunc.
	frameDone chan struct{}
}

func newGioDrawer() *GioDrawer {
	return &GioDrawer{
		frameDone: make(chan struct{}, 1),
	}
}

// Start begins a composition pass, acquiring the lock.
// The lock is held until End() is called, so Copy/Scale/Transform
// are protected from concurrent access.
func (d *GioDrawer) Start() {
	d.mu.Lock()
	d.ops.Reset()
	d.ready = false
}

// End completes the composition pass, marks ops as ready,
// submits if a frame callback is pending, and releases the lock.
func (d *GioDrawer) End() {
	d.ready = true
	d.submit()
	d.mu.Unlock()
}

func (d *GioDrawer) Redraw() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.submit()
}

// submit calls frameFunc if one is pending. Called with mu held.
// Signals frameDone after calling frameFunc so the event loop unblocks.
func (d *GioDrawer) submit() {
	if d.frameFunc != nil && d.ready {
		d.frameFunc(&d.ops)
		d.frameFunc = nil
		select {
		case d.frameDone <- struct{}{}:
		default:
		}
	}
}

// FrameReady is called by the event loop when a FrameEvent arrives.
// It stores the frame callback and blocks until the compositor calls
// End() (which calls submit), or until it can submit immediately
// because ops are already ready. If neither happens within a short
// timeout, it submits an empty frame to satisfy Gio's contract
// (e.Frame must be called before the next Event() call).
func (d *GioDrawer) FrameReady(f func(*op.Ops)) {
	d.mu.Lock()
	d.frameFunc = f
	if d.ready {
		// Ops are already built (from a previous composition), submit now.
		d.submit()
		d.mu.Unlock()
		return
	}
	d.mu.Unlock()
	// Wait for the compositor to finish, with a timeout.
	select {
	case <-d.frameDone:
	case <-time.After(50 * time.Millisecond):
		// Composition didn't complete in time. Submit an empty frame
		// so Gio doesn't try to process one itself with nil context.
		d.mu.Lock()
		if d.frameFunc != nil {
			var empty op.Ops
			d.frameFunc(&empty)
			d.frameFunc = nil
		}
		d.mu.Unlock()
	}
}

func (d *GioDrawer) Copy(dp image.Point, src image.Image, sr image.Rectangle, drawOp draw.Op, unchanged bool) {
	if sr.Empty() {
		sr = src.Bounds()
	}
	if sr.Empty() {
		return
	}
	sz := sr.Size()
	// Clip to the destination rectangle.
	stack := clip.Rect(image.Rectangle{Min: dp, Max: dp.Add(sz)}).Push(&d.ops)
	// Offset so the image is drawn at dp.
	off := op.Offset(dp.Sub(sr.Min)).Push(&d.ops)
	paint.NewImageOp(src).Add(&d.ops)
	paint.PaintOp{}.Add(&d.ops)
	off.Pop()
	stack.Pop()
}

func (d *GioDrawer) Scale(dr image.Rectangle, src image.Image, sr image.Rectangle, rotateDeg float32, drawOp draw.Op, unchanged bool) {
	if sr.Empty() {
		sr = src.Bounds()
	}
	if sr.Empty() || dr.Empty() {
		return
	}
	srcSz := sr.Size()
	dstSz := dr.Size()
	sx := float32(dstSz.X) / float32(srcSz.X)
	sy := float32(dstSz.Y) / float32(srcSz.Y)

	stack := clip.Rect(dr).Push(&d.ops)
	aff := f32.AffineId().
		Offset(f32.Pt(float32(dr.Min.X), float32(dr.Min.Y))).
		Scale(f32.Pt(0, 0), f32.Pt(sx, sy)).
		Offset(f32.Pt(float32(-sr.Min.X), float32(-sr.Min.Y)))
	if rotateDeg != 0 {
		rad := rotateDeg * (3.14159265 / 180.0)
		center := f32.Pt(float32(srcSz.X)/2, float32(srcSz.Y)/2)
		aff = aff.Rotate(center, rad)
	}
	tStack := op.Affine(aff).Push(&d.ops)
	paint.NewImageOp(src).Add(&d.ops)
	paint.PaintOp{}.Add(&d.ops)
	tStack.Pop()
	stack.Pop()
}

func (d *GioDrawer) Transform(xform math32.Matrix3, src image.Image, sr image.Rectangle, drawOp draw.Op, unchanged bool) {
	if sr.Empty() {
		sr = src.Bounds()
	}
	if sr.Empty() {
		return
	}
	// Convert math32.Matrix3 ([9]float32, column-major) to f32.Affine2D.
	// Layout: m[0]=n11(a), m[1]=n21(c), m[3]=n12(b), m[4]=n22(d), m[6]=n13(tx), m[7]=n23(ty)
	// NewAffine2D(sx, hx, ox, hy, sy, oy) = (a, b, tx, c, d, ty)
	aff := f32.NewAffine2D(
		xform[0], xform[3], xform[6],
		xform[1], xform[4], xform[7],
	)
	tStack := op.Affine(aff).Push(&d.ops)
	off := op.Offset(image.Pt(-sr.Min.X, -sr.Min.Y)).Push(&d.ops)
	paint.NewImageOp(src).Add(&d.ops)
	paint.PaintOp{}.Add(&d.ops)
	off.Pop()
	tStack.Pop()
}

func (d *GioDrawer) Renderer() any {
	return nil
}
