package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	const size = 1024
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Colors
	blue := color.RGBA{59, 130, 246, 255} // #3B82F6

	// Fill background with blue (with rounded corners)
	cornerRadius := size / 6 // ~170px for 1024
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if isInRoundedRect(x, y, size, size, cornerRadius) {
				img.Set(x, y, blue)
			}
		}
	}

	// Draw letter "C" with anti-aliasing
	centerX, centerY := float64(size)/2, float64(size)/2
	outerRadius := float64(size) * 0.35 // 358px
	strokeWidth := float64(size) * 0.12 // 123px
	innerRadius := outerRadius - strokeWidth

	// Gap angle for the "C" opening (in radians, from -45 to +45 degrees on the right)
	gapStart := -math.Pi / 4
	gapEnd := math.Pi / 4

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			// Check if point is in the ring
			if dist <= outerRadius && dist >= innerRadius {
				// Calculate angle
				angle := math.Atan2(dy, dx)

				// Skip the gap (right side opening for "C")
				if angle > gapStart && angle < gapEnd {
					continue
				}

				// Anti-aliasing at edges
				alpha := 1.0
				if dist > outerRadius-1 {
					alpha = outerRadius - dist + 1
				} else if dist < innerRadius+1 {
					alpha = dist - innerRadius + 1
				}
				if alpha > 1 {
					alpha = 1
				}
				if alpha < 0 {
					alpha = 0
				}

				c := color.RGBA{255, 255, 255, uint8(255 * alpha)}
				img.Set(x, y, c)
			}
		}
	}

	// Add sparkle effects
	sparkles := []struct {
		x, y   float64
		radius float64
		alpha  float64
	}{
		{0.78, 0.22, 0.04, 0.95},  // Large sparkle
		{0.85, 0.32, 0.025, 0.75}, // Medium sparkle
		{0.80, 0.42, 0.018, 0.55}, // Small sparkle
	}

	for _, s := range sparkles {
		sx := s.x * float64(size)
		sy := s.y * float64(size)
		sr := s.radius * float64(size)

		for y := int(sy - sr - 2); y <= int(sy+sr+2); y++ {
			for x := int(sx - sr - 2); x <= int(sx+sr+2); x++ {
				if x < 0 || x >= size || y < 0 || y >= size {
					continue
				}
				dx := float64(x) - sx
				dy := float64(y) - sy
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist <= sr {
					// Smooth edge
					alpha := s.alpha
					if dist > sr-1 {
						alpha *= (sr - dist + 1)
					}
					c := color.RGBA{255, 255, 255, uint8(255 * alpha)}
					blendPixel(img, x, y, c)
				}
			}
		}
	}

	// Save PNG to file
	f, err := os.Create("build/appicon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}

	println("Icon generated: build/appicon.png")

	// Generate ICO file for Windows
	generateICO(img)
}

func generateICO(srcImg *image.RGBA) {
	// ICO file format:
	// - Header (6 bytes)
	// - Directory entries (16 bytes each)
	// - Image data (PNG format)

	sizes := []int{256, 48, 32, 16}
	var images [][]byte

	for _, sz := range sizes {
		// Resize image
		resized := resizeImage(srcImg, sz)

		// Encode to PNG
		var buf bytes.Buffer
		png.Encode(&buf, resized)
		images = append(images, buf.Bytes())
	}

	// Create ICO file
	f, err := os.Create("build/windows/icon.ico")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Write header
	numImages := uint16(len(sizes))
	writeUint16(f, 0)         // Reserved
	writeUint16(f, 1)         // Type: 1 = ICO
	writeUint16(f, numImages) // Number of images

	// Calculate offsets
	headerSize := 6
	dirEntrySize := 16
	dataOffset := headerSize + dirEntrySize*len(sizes)

	// Write directory entries
	for i, sz := range sizes {
		width := uint8(sz)
		height := uint8(sz)
		if sz == 256 {
			width = 0 // 0 means 256
			height = 0
		}

		f.Write([]byte{width})                 // Width
		f.Write([]byte{height})                // Height
		f.Write([]byte{0})                     // Color palette
		f.Write([]byte{0})                     // Reserved
		writeUint16(f, 1)                      // Color planes
		writeUint16(f, 32)                     // Bits per pixel
		writeUint32(f, uint32(len(images[i]))) // Image size
		writeUint32(f, uint32(dataOffset))     // Image offset

		dataOffset += len(images[i])
	}

	// Write image data
	for _, imgData := range images {
		f.Write(imgData)
	}

	println("ICO generated: build/windows/icon.ico")
}

func resizeImage(src *image.RGBA, newSize int) *image.RGBA {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, newSize, newSize))

	for y := 0; y < newSize; y++ {
		for x := 0; x < newSize; x++ {
			// Bilinear interpolation
			srcX := float64(x) * float64(srcWidth) / float64(newSize)
			srcY := float64(y) * float64(srcHeight) / float64(newSize)

			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			if x1 >= srcWidth {
				x1 = srcWidth - 1
			}
			if y1 >= srcHeight {
				y1 = srcHeight - 1
			}

			xFrac := srcX - float64(x0)
			yFrac := srcY - float64(y0)

			c00 := src.RGBAAt(x0, y0)
			c10 := src.RGBAAt(x1, y0)
			c01 := src.RGBAAt(x0, y1)
			c11 := src.RGBAAt(x1, y1)

			r := bilinear(float64(c00.R), float64(c10.R), float64(c01.R), float64(c11.R), xFrac, yFrac)
			g := bilinear(float64(c00.G), float64(c10.G), float64(c01.G), float64(c11.G), xFrac, yFrac)
			b := bilinear(float64(c00.B), float64(c10.B), float64(c01.B), float64(c11.B), xFrac, yFrac)
			a := bilinear(float64(c00.A), float64(c10.A), float64(c01.A), float64(c11.A), xFrac, yFrac)

			dst.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}

	return dst
}

func bilinear(c00, c10, c01, c11, xFrac, yFrac float64) float64 {
	top := c00*(1-xFrac) + c10*xFrac
	bottom := c01*(1-xFrac) + c11*xFrac
	return top*(1-yFrac) + bottom*yFrac
}

func writeUint16(f *os.File, v uint16) {
	f.Write([]byte{byte(v), byte(v >> 8)})
}

func writeUint32(f *os.File, v uint32) {
	f.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
}

func isInRoundedRect(x, y, width, height, radius int) bool {
	if x < radius && y < radius {
		dx := radius - x
		dy := radius - y
		return dx*dx+dy*dy <= radius*radius
	}
	if x >= width-radius && y < radius {
		dx := x - (width - radius - 1)
		dy := radius - y
		return dx*dx+dy*dy <= radius*radius
	}
	if x < radius && y >= height-radius {
		dx := radius - x
		dy := y - (height - radius - 1)
		return dx*dx+dy*dy <= radius*radius
	}
	if x >= width-radius && y >= height-radius {
		dx := x - (width - radius - 1)
		dy := y - (height - radius - 1)
		return dx*dx+dy*dy <= radius*radius
	}
	return true
}

func blendPixel(img *image.RGBA, x, y int, c color.RGBA) {
	existing := img.RGBAAt(x, y)
	alpha := float64(c.A) / 255.0
	invAlpha := 1.0 - alpha

	newR := uint8(float64(c.R)*alpha + float64(existing.R)*invAlpha)
	newG := uint8(float64(c.G)*alpha + float64(existing.G)*invAlpha)
	newB := uint8(float64(c.B)*alpha + float64(existing.B)*invAlpha)
	newA := uint8(math.Min(255, float64(existing.A)+float64(c.A)))

	img.Set(x, y, color.RGBA{newR, newG, newB, newA})
}
