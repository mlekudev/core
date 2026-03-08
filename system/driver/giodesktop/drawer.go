// Copyright 2024 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giodesktop

import (
	"image"
	"image/draw"
	"sync"

	"cogentcore.org/core/math32"
	"github.com/mlekudev/gio/f32"
	"github.com/mlekudev/gio/op"
	"github.com/mlekudev/gio/op/clip"
	"github.com/mlekudev/gio/op/paint"
)

// GioDrawer implements composer.Drawer using Gio's operation stream
// to composite pre-rendered images onto the window surface.
type GioDrawer struct {
	mu sync.Mutex

	// ops is the Gio operation list built during Start/Copy/End.
	ops op.Ops

	// ready indicates that ops have been built and are ready for frame submission.
	ready bool

	// frameFunc is set by the event loop's FrameEvent handler;
	// calling it submits the ops to Gio's GPU pipeline.
	frameFunc func(*op.Ops)
}

// Start begins a composition pass, acquiring the lock.
// The lock is held until End() is called, so Copy/Scale/Transform
// are protected from concurrent SetFrameFunc access.
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
func (d *GioDrawer) submit() {
	if d.frameFunc != nil && d.ready {
		d.frameFunc(&d.ops)
		d.frameFunc = nil
	}
}

// SetFrameFunc is called by the event loop when a FrameEvent arrives.
// It stores the frame callback so that the next End() or Redraw() can submit.
func (d *GioDrawer) SetFrameFunc(f func(*op.Ops)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.frameFunc = f
	// If ops are already ready (e.g. Redraw), submit immediately.
	d.submit()
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
