package utils

import (
	"image"
	"math"
)

const (
	KERNEL    = 3
	Threshold = 16
)

func kern(pic *image.RGBA64, x, y int) (r, g, b, a uint32) {
	for i := 0; i <= KERNEL; i++ {
		for j := 0; j <= KERNEL; j++ {
			var c = pic.RGBA64At(x+i, y+j)
			r0, g0, b0, a0 := c.RGBA()
			r += r0
			g += g0
			b += b0
			a += a0
		}
	}
	r >>= 12
	g >>= 12
	b >>= 12
	a >>= 12
	return r, g, b, a
}

// Compare16 ...
func Compare16(pic1 *image.RGBA64, pic2 *image.RGBA64) (stats DiffStats, err error) {
	stats.ExactSame = true
	stats.NumPixels = int64(pic1.Bounds().Dx()) * int64(pic1.Bounds().Dy())
	for y := 0; y < (pic1.Bounds().Dy() - KERNEL); y++ {
		for x := 0; x < (pic1.Bounds().Dx() - KERNEL); x++ {
			r1, g1, b1, a1 := kern(pic1, x, y)
			r2, g2, b2, a2 := kern(pic2, x, y)

			var diff = uint32(math.Abs(float64(r1) - float64(r2)))
			diff += uint32(math.Abs(float64(g1) - float64(g2)))
			diff += uint32(math.Abs(float64(b1) - float64(b2)))
			diff += uint32(math.Abs(float64(a1) - float64(a2)))

			if diff != 0 {
				stats.ExactSame = false
			}

			diff >>= 2
			same := diff <= Threshold

			if !same {
				stats.DiffPixels++
			}
		}
	}

	return stats, nil
}
