package main

import (
	//"fmt"

	"github.com/joshvictor1024/go-mandelbrot/pkg/types"
)

const (
	CHUNK_LENGTH int = 128
	//EPSILON float64 =
)

type chunk struct {
	originNumber   types.Pointf64
	numberPerTexel float64
}

func (ch *chunk) good(originNumber types.Pointf64, numberPerTexel float64) bool {
	epsilon := numberPerTexel / 100
	dx := ch.originNumber.X - originNumber.X
	if dx > epsilon || dx+epsilon < 0 {
		return false
	}
	dy := ch.originNumber.Y - originNumber.Y
	if dy > epsilon || dy+epsilon < 0 {
		return false
	}
	dnpt := ch.numberPerTexel - numberPerTexel
	if dnpt > epsilon || dnpt+epsilon < 0 {
		return false
	}
	return true
}

type iterateWork struct {
	originNumber   types.Pointf64
	numberPerTexel float64
	*chunk
	*iterationBuffer
	*drawWork
}

type drawWork struct {
	textureData       []byte
	textureDataWidth  int
	textureDataOrigin types.Pointi
	chunk             *chunk
	*iterationBuffer
}

type iterationBuffer struct {
	data  [CHUNK_LENGTH][CHUNK_LENGTH]uint // [y][x]
	inUse bool
}

func iterateChunk(iw *iterateWork) {
	data := &iw.iterationBuffer.data
	for yi := 0; yi < CHUNK_LENGTH; yi += 1 {
		for xi := 0; xi < CHUNK_LENGTH; xi += 1 {
			data[yi][xi] = iterate(
				iw.originNumber.X+float64(xi)*iw.numberPerTexel,
				iw.originNumber.Y-float64(yi)*iw.numberPerTexel,
				255,
			)
		}
	}
}

func drawChunk(dw *drawWork) {
	for yi := 0; yi < CHUNK_LENGTH; yi += 1 {
		for xi := 0; xi < CHUNK_LENGTH; xi += 1 {
			pixel := (dw.textureDataOrigin.Y+yi)*dw.textureDataWidth + (dw.textureDataOrigin.X + xi)
			color := byte(dw.iterationBuffer.data[yi][xi])
			dw.textureData[pixel*4+3] = color
			dw.textureData[pixel*4+2] = color
			dw.textureData[pixel*4+1] = color
			dw.textureData[pixel*4+0] = 255
		}
	}
}
