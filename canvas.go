package main

import (
	//"math"
	"fmt"
	"sync"

	"github.com/joshvictor1024/go-mandelbrot/pkg/types"
	"github.com/veandco/go-sdl2/sdl"
)

type canvas struct {
	renderer       *sdl.Renderer
	texture        *sdl.Texture
	chunks         [][]chunk // [y][x], updated to latest
	originChunk    types.Pointi
	originNumber   types.Pointf64
	numberPerTexel float64
	iq             *iterateQueue
	dq             *drawQueue
	ibs            *iterationBufferQueue
	stopCh         chan struct{}
}

const (
	IBS_CAP   = 1
	IT_WORKER = 1
)

func newCanvas(r *sdl.Renderer, minWidth, minHeight int, numberOrigin types.Pointf64, numberPerTexel float64) *canvas {
	chunkW := minWidth/CHUNK_LENGTH + 3
	chunkH := minHeight/CHUNK_LENGTH + 3

	t, err := r.CreateTexture(
		sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_STREAMING, // only textures with TEXTUREACCESS_STREAMING can be locked
		int32(chunkW*CHUNK_LENGTH),
		int32(chunkH*CHUNK_LENGTH),
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	chunks := make([][]chunk, chunkH)
	for y := 0; y < chunkH; y += 1 {
		chunks[y] = make([]chunk, 0, chunkW)
		for x := 0; x < chunkW; x += 1 {
			chunks[y] = append(chunks[y], chunk{})
		}
	}

	dirty := map[*chunk]struct{}{}
	iq := newIterateQueue(dirty)
	dq := newDrawQueue()
	ibs := newIterationBufferQueue()
	for i := 0; i < IBS_CAP; i += 1 {
		ibs.send(&iterationBuffer{})
	}

	return &canvas{
		renderer:       r,
		texture:        t,
		chunks:         chunks,
		originNumber:   numberOrigin,
		numberPerTexel: numberPerTexel,
		iq:             iq,
		dq:             dq,
		ibs:            ibs,
		stopCh:         make(chan struct{}),
	}
}

func (c *canvas) close() {
	fmt.Println("closing stopCh...")
	close(c.stopCh)
	fmt.Println("closing other channels...")
	c.iq.close()
	c.dq.close()
	c.ibs.close()
	fmt.Println("destroying texture...")
	c.texture.Destroy()
	fmt.Println("done destroy texture")
}

// blocks until stopCh is closed
// call this in a go routine
func (c *canvas) work() {
	//fmt.Println("enter work")
	wg := new(sync.WaitGroup)
	defer wg.Wait()
	wg.Add(8)
	for i := 0; i < IT_WORKER; i += 1 {
		go c.processIterationWork(wg)
	}
	<-c.stopCh
}

func (c *canvas) processIterationWork(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		fmt.Println("attempt getting from c.iq...")
		iw, ok := c.iq.recv()
		fmt.Println("got from c.iq")
		if ok == false {
			fmt.Println("failed to get from iq")
			return
		}
		var ib *iterationBuffer = nil
		for {
			//fmt.Println("attempt getting from c.ibs...")
			//fmt.Printf("ibs has %v items\n", c.ibs.len())
			var ok bool
			ib, ok = c.ibs.recv()
			if ok == false {
				fmt.Println("failed get from ibs")
				return
			}

			if ib.inUse == false {
				//fmt.Println("got from c.ibs")
				break
			}

			//fmt.Println("attempt pushing to c.ibs...")
			if c.ibs.send(ib) == false {
				fmt.Println("failed pushing to ibs")
				return
			}
			//fmt.Println("pushed to c.ibs")
		}
		ib.inUse = true
		iw.iterationBuffer = ib

		iterateChunk(iw)
		iw.drawWork.iterationBuffer = ib

		fmt.Println("attempt pushing to dq...")
		if c.dq.send(iw.drawWork) == false {
			fmt.Println("failed pushing to dq")
			return
		}
		fmt.Println("pushed to c.dq")
	}
}

func (c *canvas) generate() {
	fmt.Println("generate()")
	for yi := 0; yi < c.getHeightChunkCount(); yi += 1 {
		for xi := 0; xi < c.getWidthChunkCount(); xi += 1 {
			chunk := &(c.chunks[yi][xi])

			originNumber := types.Pointf64{
				X: c.originNumber.X + float64(xi*CHUNK_LENGTH)*c.numberPerTexel,
				Y: c.originNumber.Y - float64(yi*CHUNK_LENGTH)*c.numberPerTexel,
			}
			if chunk.good(originNumber, c.numberPerTexel) {
				//fmt.Println("good")
				continue
			}
			chunk.originNumber = originNumber
			chunk.numberPerTexel = c.numberPerTexel

			textureDataOrigin := types.Pointi{
				X: c.originChunk.X + xi,
				Y: c.originChunk.Y + yi,
			}
			if textureDataOrigin.X >= c.getWidthChunkCount() {
				textureDataOrigin.X -= c.getWidthChunkCount()
			}
			if textureDataOrigin.Y >= c.getHeightChunkCount() {
				textureDataOrigin.Y -= c.getHeightChunkCount()
			}
			textureDataOrigin.X *= CHUNK_LENGTH
			textureDataOrigin.Y *= CHUNK_LENGTH

			if c.iq.send(&iterateWork{
				chunk:          chunk,
				originNumber:   originNumber,
				numberPerTexel: c.numberPerTexel,
				drawWork: &drawWork{
					textureDataWidth:  c.getWidthTexel(),
					textureDataOrigin: textureDataOrigin,
					chunk:             chunk,
				},
			}) == false {
				return
			}

			// fmt.Printf("generating chunk (%v, %v) indexed (%v, %v) at number (%v, %v)\n",
			// 	xi, yi, dataOrigin.x, dataOrigin.y, numberOrigin.x, numberOrigin.y,
			// )
		}
	}
}

func (c *canvas) draw(numberRect types.Rectf64) {
	fmt.Println("draw()")
	data, _, err := c.texture.Lock(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		canRecv, dw, ok := c.dq.attemptRecv(false)
		if ok == false {
			return
		}
		if canRecv == false {
			break
		}

		dw.textureData = data
		//fmt.Printf("dw: %v %v %v\n", dw.textureDataWidth, dw.textureDataOrigin, dw.chunk)
		drawChunk(dw)
		dw.iterationBuffer.inUse = false
		if c.ibs.send(dw.iterationBuffer) == false {
			return
		}
	}
	//fmt.Println("done drawChunk")
	c.texture.Unlock()

	tr := c.toTextureRecti(numberRect)
	xn, xp, yn, yp := c.checkTextureRectiBounds(tr)
	if xn {
		fmt.Println("out of bounds: xn")
	}
	if xp {
		fmt.Println("out of bounds: xp")
	}
	if yn {
		fmt.Println("out of bounds: yn")
	}
	if yp {
		fmt.Println("out of bounds: yp")
	}
	if xn || xp || yn || yp {
		return
	}

	//fmt.Printf("textureRect: %v\n", textureRect)
	c.renderer.Copy(c.texture, &sdl.Rect{X: int32(tr.X), Y: int32(tr.Y), W: int32(tr.W), H: int32(tr.H)}, nil)
	//fmt.Println("done draw")
}

func (c *canvas) dump(numberRect types.Rectf64, dst types.Pointf64, ratio float32) {
	r := sdl.FRect{
		X: float32(dst.X),
		Y: float32(dst.Y),
		W: float32(c.getWidthTexel()) * ratio,
		H: float32(c.getHeightTexel()) * ratio,
	}
	c.renderer.CopyF(c.texture, nil, &r)
	c.renderer.DrawRectF(&r)

	tr := c.toTextureRecti(numberRect)
	xn, xp, yn, yp := c.checkTextureRectiBounds(tr)
	if xn {
		fmt.Println("out of bounds: xn")
	}
	if xp {
		fmt.Println("out of bounds: xp")
	}
	if yn {
		fmt.Println("out of bounds: yn")
	}
	if yp {
		fmt.Println("out of bounds: yp")
	}
	if xn || xp || yn || yp {
		return
	}

	c.renderer.DrawRect(&sdl.Rect{
		X: int32(float32(dst.X) + (float32(tr.X) * ratio)),
		Y: int32(float32(dst.Y) + (float32(tr.Y) * ratio)),
		W: int32(float32(tr.W) * ratio),
		H: int32(float32(tr.H) * ratio),
	})
}

func (c *canvas) toTextureRecti(numberRect types.Rectf64) *types.Recti {
	//fmt.Printf("numberRect: %v\n", numberRect)

	textureDx := int((numberRect.X - c.originNumber.X) / c.numberPerTexel)
	textureDy := int((numberRect.Y - c.originNumber.Y) / c.numberPerTexel)
	//fmt.Printf("%v %v\n", textureDx, textureDy)

	textureW := int(numberRect.W / c.numberPerTexel)
	textureH := int(numberRect.H / c.numberPerTexel)
	//fmt.Printf("%v %v\n", textureW, textureH)
	//fmt.Printf("%v %v\n", c.getChunkCountX(), c.getChunkCountY())

	return &types.Recti{
		X: c.getOriginTexelX() + textureDx,
		Y: c.getOriginTexelY() + textureDy,
		W: textureW,
		H: textureH,
	}
}

func (c *canvas) checkTextureRectiBounds(textureRect *types.Recti) (
	xNeg, xPos, yNeg, yPos bool,
) {
	ox := c.getOriginTexelX()
	xNeg = textureRect.X < ox
	xPos = textureRect.X+textureRect.W > ox+c.getWidthTexel()
	oy := c.getOriginTexelY()
	yNeg = textureRect.Y < oy
	yPos = textureRect.Y+textureRect.H > oy+c.getWidthTexel()
	return
}

func (c *canvas) getWidthTexel() int {
	return c.getWidthChunkCount() * CHUNK_LENGTH
}

func (c *canvas) getHeightTexel() int {
	return c.getHeightChunkCount() * CHUNK_LENGTH
}

func (c *canvas) getHeightChunkCount() int {
	return len(c.chunks)
}

func (c *canvas) getWidthChunkCount() int {
	return len(c.chunks[0])
}

func (c *canvas) getOriginTexelX() int {
	return c.originChunk.X * CHUNK_LENGTH
}

func (c *canvas) getOriginTexelY() int {
	return c.originChunk.Y * CHUNK_LENGTH
}
