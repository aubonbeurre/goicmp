package utils

import (
	"image"
	"math"
)

// Compare16 ...
func Compare16(pic1 *image.RGBA64, pic2 *image.RGBA64) (stats DiffStats, err error) {
	stats.ExactSame = true
	stats.NumPixels = int64(pic1.Bounds().Dx()) * int64(pic1.Bounds().Dy())
	var KERNEL = 3
	var Threshold uint32 = 16
	for y := 0; y < (pic1.Bounds().Dy() - KERNEL); y++ {
		for x := 0; x < (pic1.Bounds().Dx() - KERNEL); x++ {
			var r1, g1, b1, a1 uint32
			r1, g1, b1, a1 = 0, 0, 0, 0

			for i := 0; i <= KERNEL; i++ {
				for j := 0; j <= KERNEL; j++ {
					var c = pic1.RGBA64At(x+i, y+i)
					r0, g0, b0, a0 := c.RGBA()
					r1 += r0
					g1 += g0
					b1 += b0
					a1 += a0
				}
			}

			r1 >>= 12
			g1 >>= 12
			b1 >>= 12
			a1 >>= 12

			var r2, g2, b2, a2 uint32
			r2, g2, b2, a2 = 0, 0, 0, 0

			for i := 0; i <= KERNEL; i++ {
				for j := 0; j <= KERNEL; j++ {
					var c = pic2.RGBA64At(x+i, y+i)
					r0, g0, b0, a0 := c.RGBA()
					r2 += r0
					g2 += g0
					b2 += b0
					a2 += a0
				}
			}

			r2 >>= 12
			g2 >>= 12
			b2 >>= 12
			a2 >>= 12

			var diff = uint32(math.Abs(float64(r1) - float64(r2)))
			diff += uint32(math.Abs(float64(g1) - float64(g2)))
			diff += uint32(math.Abs(float64(b1) - float64(b2)))
			diff += uint32(math.Abs(float64(a1) - float64(a2)))

			diff >>= 2
			same := diff <= Threshold

			if !same {
				stats.ExactSame = false
				stats.DiffPixels++
			}
		}
	}

	return stats, nil
}
