package main

import (
	//"fmt"

	"github.com/joshvictor1024/go-mandelbrot/pkg/types"
	"github.com/veandco/go-sdl2/sdl"
)

type scene struct {
	renderer       *sdl.Renderer
	canvas         *canvas
	w              int32
	h              int32
	numberPerPixel float64
	numberOrigin   types.Pointf64
}

func newScene(r *sdl.Renderer, w, h int32) *scene {
	s := scene{
		renderer:       r,
		w:              w,
		h:              h,
		numberPerPixel: 0.003,
		numberOrigin:   types.Pointf64{X: -2, Y: 1},
	}
	c := newCanvas(r, int(w), int(h), s.numberOrigin, s.numberPerPixel)
	if c == nil {
		return nil
	}
	s.canvas = c
	go c.work()
	return &s
}

func (s *scene) close() {
	s.canvas.close()
}

func (s *scene) draw() {
	s.canvas.generate()

	numberRect := types.Rectf64{
		X: s.numberOrigin.X,
		Y: s.numberOrigin.Y,
		W: float64(s.w) * s.numberPerPixel,
		H: float64(s.h) * s.numberPerPixel,
	}
	s.canvas.draw(numberRect)
	s.canvas.dump(numberRect, types.Pointf64{X: 0, Y: 0}, 0.15)
}

func (s *scene) updateView(deltaPixel types.Pointi, ratioNumberPerPixel float64) {
	//fmt.Printf("(%v, %v) %v\n", deltaPixel.x, deltaPixel.y, ratioNumberPerPixel)
	s.numberOrigin.X += float64(deltaPixel.X) * s.numberPerPixel
	s.numberOrigin.Y += float64(deltaPixel.Y) * s.numberPerPixel
}
