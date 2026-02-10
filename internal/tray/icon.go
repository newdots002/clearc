package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

// GetIconData returns the icon data for the system tray
// This generates a 32x32 ICO icon with blue background (#3B82F6) and white "C" letter
func GetIconData() []byte {
	const size = 32
	img := createIconImage(size)

	// Convert to ICO format for Windows
	return createICO(img, size)
}

func createIconImage(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Colors
	blue := color.RGBA{59, 130, 246, 255}   // #3B82F6
	white := color.RGBA{255, 255, 255, 255} // White

	// Fill background with blue (with rounded corners effect)
	cornerRadius := size / 5
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if isInRoundedRect(x, y, size, size, cornerRadius) {
				img.Set(x, y, blue)
			}
		}
	}

	// Draw letter "C" - using a simple arc approximation
	centerX, centerY := size/2, size/2
	outerRadius := size * 10 / 32
	innerRadius := size * 6 / 32

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - centerX
			dy := y - centerY
			dist := dx*dx + dy*dy

			// Check if point is in the ring (between inner and outer radius)
			if dist <= outerRadius*outerRadius && dist >= innerRadius*innerRadius {
				// Create the "C" opening on the right side
				gapSize := size * 4 / 32
				if x > centerX && dy > -gapSize && dy < gapSize {
					continue
				}
				img.Set(x, y, white)
			}
		}
	}

	// Add small sparkle dots
	sparkles := []struct{ x, y int }{
		{size * 25 / 32, size * 8 / 32},
		{size * 27 / 32, size * 11 / 32},
		{size * 26 / 32, size * 14 / 32},
	}
	for _, s := range sparkles {
		if s.x < size && s.y < size {
			img.Set(s.x, s.y, white)
		}
	}

	return img
}

// createICO creates an ICO file format from an image
func createICO(img *image.RGBA, size int) []byte {
	var buf bytes.Buffer

	// Encode image as PNG first
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	pngData := pngBuf.Bytes()

	// ICO Header (6 bytes)
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // Reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // Type: 1 = ICO
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // Number of images

	// ICO Directory Entry (16 bytes)
	iconWidth := uint8(size)
	iconHeight := uint8(size)
	if size == 256 {
		iconWidth = 0
		iconHeight = 0
	}
	buf.WriteByte(iconWidth)                                      // Width
	buf.WriteByte(iconHeight)                                     // Height
	buf.WriteByte(0)                                              // Color palette
	buf.WriteByte(0)                                              // Reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1))            // Color planes
	binary.Write(&buf, binary.LittleEndian, uint16(32))           // Bits per pixel
	binary.Write(&buf, binary.LittleEndian, uint32(len(pngData))) // Image size
	binary.Write(&buf, binary.LittleEndian, uint32(22))           // Offset (6 header + 16 directory)

	// PNG data
	buf.Write(pngData)

	return buf.Bytes()
}

// isInRoundedRect checks if a point is inside a rounded rectangle
func isInRoundedRect(x, y, width, height, radius int) bool {
	// Check corners
	if x < radius && y < radius {
		// Top-left corner
		dx := radius - x
		dy := radius - y
		return dx*dx+dy*dy <= radius*radius
	}
	if x >= width-radius && y < radius {
		// Top-right corner
		dx := x - (width - radius - 1)
		dy := radius - y
		return dx*dx+dy*dy <= radius*radius
	}
	if x < radius && y >= height-radius {
		// Bottom-left corner
		dx := radius - x
		dy := y - (height - radius - 1)
		return dx*dx+dy*dy <= radius*radius
	}
	if x >= width-radius && y >= height-radius {
		// Bottom-right corner
		dx := x - (width - radius - 1)
		dy := y - (height - radius - 1)
		return dx*dx+dy*dy <= radius*radius
	}
	return true
}
