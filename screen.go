package main

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenW     = 64
	screenH     = 32
	scaleFactor = 10
	sampleRate  = 48000
	frequency   = 440
)

type Debug struct {
	enabled bool
	step    bool
}

type Display struct {
	frameBuffer []byte
	img         *ebiten.Image
	delayTimer  byte
	soundTimer  byte

	keys []bool

	debug Debug

	audioContext *audio.Context
	audioPlayer  *audio.Player
}

type stream struct {
	pos int64
}

func (s *stream) Read(buf []byte) (int, error) {
	const bytesPerSample = 8

	n := (len(buf) / bytesPerSample * bytesPerSample)

	const length = sampleRate / frequency
	for i := 0; i < n/bytesPerSample; i++ {
		v := math.Float32bits(float32(math.Sin(2 * math.Pi * float64(s.pos/bytesPerSample+int64(i)) / length)))
		buf[8*i] = byte(v)
		buf[8*i+1] = byte(v >> 8)
		buf[8*i+2] = byte(v >> 16)
		buf[8*i+3] = byte(v >> 24)
		buf[8*i+4] = byte(v)
		buf[8*i+5] = byte(v >> 8)
		buf[8*i+6] = byte(v >> 16)
		buf[8*i+7] = byte(v >> 24)
	}

	s.pos += int64(n)
	s.pos %= length * bytesPerSample

	return n, nil
}

func (s *stream) Close() error {
	return nil
}

func NewDisplay() *Display {
	return &Display{
		frameBuffer: make([]byte, screenW*screenH*4),
		img:         ebiten.NewImage(screenW, screenH),
		delayTimer:  0,
		soundTimer:  0,
		keys:        make([]bool, 16),
	}
}

func (d *Display) EnableDebug() {
	d.debug.enabled = true
}

func (d *Display) DisableDebug() {
	d.debug.enabled = false
}

func (d *Display) ClearScreen() {
	for i := range screenW * screenH * 4 {
		d.frameBuffer[i] = 0
	}
}

func (d *Display) DrawSprite(c *Cpu, spriteData byte, x uint8, y uint8) {
	x_coord := uint16(x)
	y_coord := uint16(y)
	idx := ((y_coord%32)*64 + (x_coord % 64)) * 4
	c.regs[0xF] = 0
	for i := range 8 {
		if x_coord+uint16(i) > 63 {
			break
		}
		//0b11011001
		//(0b11011001 & 1 << (8 - i - 1)) >> (8 - i - 1)
		pixel := spriteData & (1 << (7 - i)) >> (7 - i)
		if (pixel*255)&d.frameBuffer[idx+0+uint16(i)*4] == 255 {
			c.regs[0xF] = 1
		}
		d.frameBuffer[idx+0+uint16(i)*4] = (pixel * 255) ^ d.frameBuffer[idx+0+uint16(i)*4]
		d.frameBuffer[idx+1+uint16(i)*4] = (pixel * 255) ^ d.frameBuffer[idx+1+uint16(i)*4]
		d.frameBuffer[idx+2+uint16(i)*4] = (pixel * 255) ^ d.frameBuffer[idx+2+uint16(i)*4]
		d.frameBuffer[idx+3+uint16(i)*4] = 255
	}
}

func (d *Display) Update() error {
	if d.audioContext == nil {
		d.audioContext = audio.NewContext(sampleRate)
	}
	if d.audioPlayer == nil {
		var err error
		d.audioPlayer, err = d.audioContext.NewPlayerF32(&stream{})
		if err != nil {
			return err
		}
	}
	d.img.WritePixels(d.frameBuffer)
	if d.delayTimer > 0 {
		d.delayTimer--
	}
	if d.soundTimer > 0 {
		if !d.audioPlayer.IsPlaying() {
			d.audioPlayer.Play()
		}
		d.soundTimer--
	} else if d.soundTimer == 0 {
		d.audioPlayer.Pause()
	}
	var keys = []ebiten.Key{
		ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4,
		ebiten.KeyQ, ebiten.KeyW, ebiten.KeyE, ebiten.KeyR,
		ebiten.KeyA, ebiten.KeyS, ebiten.KeyD, ebiten.KeyF,
		ebiten.KeyZ, ebiten.KeyX, ebiten.KeyC, ebiten.KeyV,
	}
	for i := range len(keys) {
		if ebiten.IsKeyPressed(keys[i]) {
			d.keys[i] = true
		} else {
			d.keys[i] = false
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		d.debug.step = true
	} else {
		d.debug.step = false
	}
	return nil
}

func (d *Display) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleFactor, scaleFactor)
	screen.DrawImage(d.img, op)
}

func (d *Display) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW * scaleFactor, screenH * scaleFactor
}

func RunDisplay(display *Display) {
	ebiten.SetWindowSize(screenW*scaleFactor, screenH*scaleFactor)
	ebiten.SetWindowTitle("CHIP-8 emulator by Dominik Mészáros")
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(display); err != nil {
		log.Fatal(err)
	}
}
