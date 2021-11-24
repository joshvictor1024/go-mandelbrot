# go-mandelbrot

A mandelbrot set fractal generator, demonstrating parallelism using go's
concurrency features

## Demo

### Build

under root run `go build`

### App

- Drag with left mouse button to pan
- Press Esc to exit

## Planned Feature

- Basic panning and zooming
- Skip chunks when not dirty
- Parallelism with work queue item compressing
- Pre-generate chunks
- Support arbitrary precision (infinite zoom)
- Decouple from SDL2 for rendering

## Implementation

### Overview

- `scene` does drawing, and transforms pixel-space actions of the user to
  number-space.
- `canvas` abstracts over hardware texture, and transform number-space actions
  from `scene` to texel-space. handles when to do iteration and draw work. owns
  workqueues that manages `chunk`s to be processed.
- `chunk` regions in texel space. smallest unit on which work is done.

### Parallelizing Work

Computing iterations can be trivially parallelized since it only depends on
local data i.e. the represented complex numbers within the chunk. Also, since
it's quite long-running, it greatly benefits from parallelism.

On the other hand, the speed of drawing to texture is bound by uploading data
to the GPU, so running on a single thread is sufficient.