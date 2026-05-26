package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

var iconSizes = []int{16, 32, 48, 64, 128, 256}

func main() {
	must(os.MkdirAll(filepath.Join("assets", "logo"), 0o755))
	for _, size := range iconSizes {
		img := drawIcon(size)
		out, err := os.Create(filepath.Join("assets", "logo", "voicecast-"+itoa(size)+".png"))
		must(err)
		must(png.Encode(out, img))
		must(out.Close())
	}
	must(writeICO(filepath.Join("assets", "logo", "voicecast.ico")))
}

func drawIcon(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	radius := float64(size) * 0.205
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if roundedInside(float64(x)+0.5, float64(y)+0.5, float64(size), radius) {
				img.SetRGBA(x, y, color.RGBA{R: 0x10, G: 0x18, B: 0x20, A: 0xff})
			}
		}
	}

	fillPoly(img, []point{
		{0.17, 0.36}, {0.37, 0.36}, {0.61, 0.20}, {0.61, 0.80}, {0.37, 0.64}, {0.17, 0.64},
	}, size, func(x, y int) color.RGBA {
		return grad(x, y, size)
	})

	strokeArc(img, size, 0.63, 0.50, 0.15, -0.70, 0.70, 0.055, color.RGBA{R: 0xf7, G: 0xff, B: 0xfb, A: 0xff})
	strokeArc(img, size, 0.66, 0.50, 0.25, -0.78, 0.78, 0.045, color.RGBA{R: 0xf7, G: 0xff, B: 0xfb, A: 0xdb})
	strokeLine(img, size, point{0.14, 0.15}, point{0.86, 0.85}, 0.035, color.RGBA{R: 0xc7, G: 0xf6, B: 0xff, A: 0x45})

	return img
}

func roundedInside(x, y, size, radius float64) bool {
	min := radius
	max := size - radius
	cx := clamp(x, min, max)
	cy := clamp(y, min, max)
	dx := x - cx
	dy := y - cy
	return dx*dx+dy*dy <= radius*radius
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

type point struct {
	x float64
	y float64
}

func fillPoly(img *image.RGBA, pts []point, size int, c func(int, int) color.RGBA) {
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if pointInPoly((float64(x)+0.5)/float64(size), (float64(y)+0.5)/float64(size), pts) {
				img.SetRGBA(x, y, c(x, y))
			}
		}
	}
}

func pointInPoly(x, y float64, pts []point) bool {
	inside := false
	j := len(pts) - 1
	for i := range pts {
		if (pts[i].y > y) != (pts[j].y > y) && x < (pts[j].x-pts[i].x)*(y-pts[i].y)/(pts[j].y-pts[i].y)+pts[i].x {
			inside = !inside
		}
		j = i
	}
	return inside
}

func grad(x, y, size int) color.RGBA {
	t := float64(x+y) / float64(size*2)
	if t < 0.55 {
		p := t / 0.55
		return lerp(color.RGBA{0x1f, 0xb6, 0xa6, 0xff}, color.RGBA{0x2e, 0x7d, 0xd7, 0xff}, p)
	}
	p := (t - 0.55) / 0.45
	return lerp(color.RGBA{0x2e, 0x7d, 0xd7, 0xff}, color.RGBA{0x5c, 0x6b, 0xc0, 0xff}, p)
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: uint8(float64(a.A) + (float64(b.A)-float64(a.A))*t),
	}
}

func strokeArc(img *image.RGBA, size int, cx, cy, r, a1, a2, width float64, col color.RGBA) {
	steps := int(float64(size) * 1.5)
	prev := point{cx + math.Cos(a1)*r, cy + math.Sin(a1)*r}
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		a := a1 + (a2-a1)*t
		next := point{cx + math.Cos(a)*r, cy + math.Sin(a)*r}
		strokeLine(img, size, prev, next, width, col)
		prev = next
	}
}

func strokeLine(img *image.RGBA, size int, a, b point, width float64, col color.RGBA) {
	ax, ay := a.x*float64(size), a.y*float64(size)
	bx, by := b.x*float64(size), b.y*float64(size)
	w := width * float64(size)
	minX := int(math.Max(0, math.Floor(math.Min(ax, bx)-w)))
	maxX := int(math.Min(float64(size-1), math.Ceil(math.Max(ax, bx)+w)))
	minY := int(math.Max(0, math.Floor(math.Min(ay, by)-w)))
	maxY := int(math.Min(float64(size-1), math.Ceil(math.Max(ay, by)+w)))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			d := lineDistance(float64(x)+0.5, float64(y)+0.5, ax, ay, bx, by)
			if d <= w/2 {
				blend(img, x, y, col)
			}
		}
	}
}

func lineDistance(px, py, ax, ay, bx, by float64) float64 {
	dx := bx - ax
	dy := by - ay
	if dx == 0 && dy == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
	t = clamp(t, 0, 1)
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}

func blend(img *image.RGBA, x, y int, src color.RGBA) {
	dst := img.RGBAAt(x, y)
	a := float64(src.A) / 255
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(float64(src.R)*a + float64(dst.R)*(1-a)),
		G: uint8(float64(src.G)*a + float64(dst.G)*(1-a)),
		B: uint8(float64(src.B)*a + float64(dst.B)*(1-a)),
		A: uint8(math.Min(255, float64(src.A)+float64(dst.A)*(1-a))),
	})
}

func writeICO(path string) error {
	type frame struct {
		size int
		data []byte
	}
	var frames []frame
	for _, size := range iconSizes {
		var buf bytes.Buffer
		if err := png.Encode(&buf, drawIcon(size)); err != nil {
			return err
		}
		frames = append(frames, frame{size: size, data: buf.Bytes()})
	}
	var out bytes.Buffer
	binary.Write(&out, binary.LittleEndian, uint16(0))
	binary.Write(&out, binary.LittleEndian, uint16(1))
	binary.Write(&out, binary.LittleEndian, uint16(len(frames)))
	offset := 6 + len(frames)*16
	for _, f := range frames {
		w := byte(f.size)
		if f.size == 256 {
			w = 0
		}
		out.WriteByte(w)
		out.WriteByte(w)
		out.WriteByte(0)
		out.WriteByte(0)
		binary.Write(&out, binary.LittleEndian, uint16(1))
		binary.Write(&out, binary.LittleEndian, uint16(32))
		binary.Write(&out, binary.LittleEndian, uint32(len(f.data)))
		binary.Write(&out, binary.LittleEndian, uint32(offset))
		offset += len(f.data)
	}
	for _, f := range frames {
		out.Write(f.data)
	}
	return os.WriteFile(path, out.Bytes(), 0o644)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
