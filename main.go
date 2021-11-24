package main

import (
	"fmt"
	// "math/rand"
	// "time"

	"github.com/joshvictor1024/go-mandelbrot/pkg/types"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowWidth  = 800
	windowHeight = 600
)

func sdlInit(windowTitle string) (*sdl.Window, *sdl.Renderer, error) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_TIMER); err != nil {
		return nil, nil, err
	}
	sdl.StopTextInput()

	window, err := sdl.CreateWindow(
		windowTitle,
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		windowWidth, windowHeight, sdl.WINDOW_OPENGL,
	)
	if err != nil {
		return nil, nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, nil, err
	}

	return window, renderer, nil
}

func sdlClose(window *sdl.Window, renderer *sdl.Renderer) {
	renderer.Destroy()
	window.Destroy()
	sdl.Quit()
}

func main() {
	// start SDL
	window, renderer, err := sdlInit("Mandelbrot")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdlClose(window, renderer)

	s := newScene(renderer, int32(windowWidth), int32(windowHeight))
	defer s.close()

	// start loop
	run := true
	mouseDownPosition := types.Pointi{}

	for run {
		// WaitEvent must be on the same thread that did INIT_VIDEO
		e := sdl.WaitEvent()

		// WaitEvent returns nil on some error
		if e == nil {
			fmt.Print("event is nil!\n")
			run = false
			break
		}

		switch t := e.(type) {
		case *sdl.QuitEvent:
			fmt.Print("quit event\n")
			run = false
		case *sdl.MouseButtonEvent:
			if t.Type == sdl.MOUSEBUTTONDOWN {
				mouseDownPosition = types.Pointi{X: int(t.X), Y: int(t.Y)}
			} else if t.Type == sdl.MOUSEBUTTONUP {
				mousePositionDelta := types.Pointi{
					X: mouseDownPosition.X - int(t.X),
					Y: mouseDownPosition.Y - int(t.Y),
				}
				s.updateView(mousePositionDelta, 1)
			}
		case *sdl.KeyboardEvent:
			if t.Keysym.Sym == sdl.K_ESCAPE {
				fmt.Print("esc event\n")
				run = false
			}
			// default:
			// 	fmt.Printf("Event: %T\n", t)
		}

		// draw
		renderer.SetDrawColor(255, 0, 255, 255)
		renderer.Clear()
		s.draw()
		renderer.Present()
	}
}
