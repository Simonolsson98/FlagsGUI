package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand"
	"net/http"
)

// downloadFlagImage downloads a flag image from the given URL
func downloadFlagImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the image data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// getDistinctColors extracts dominant colors from flag image, filtering out noise and anti-aliasing artifacts
func getDistinctColors(img image.Image) []color.RGBA {
	bounds := img.Bounds()
	colorMap := make(map[[3]uint8]int) // RGB only, ignore alpha variations

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			// Skip transparent/semi-transparent pixels
			if uint8(a>>8) < 128 {
				continue
			}

			// Quantize colors to reduce anti-aliasing variations (round to nearest 16)
			// This groups similar colors together (e.g., RGB(200,45,50) and RGB(205,48,52) become the same)
			rgb := [3]uint8{
				(uint8(r>>8) / 16) * 16,
				(uint8(g>>8) / 16) * 16,
				(uint8(b>>8) / 16) * 16,
			}

			colorMap[rgb]++
		}
	}

	// Calculate threshold: only return colors that appear in at least 0.5% of pixels
	totalPixels := (bounds.Max.X - bounds.Min.X) * (bounds.Max.Y - bounds.Min.Y)
	threshold := totalPixels / 200 // 0.5% threshold
	if threshold < 10 {
		threshold = 10 // Minimum threshold for small images
	}

	var colors []color.RGBA
	for rgb, count := range colorMap {
		if count > threshold {
			colors = append(colors, color.RGBA{
				R: rgb[0],
				G: rgb[1],
				B: rgb[2],
				A: 255,
			})
		}
	}

	return colors
}

// colorsAreSimilar checks if two colors are similar within a given tolerance
func colorsAreSimilar(c1, c2 color.RGBA, tolerance uint8) bool {
	diffR := int(c1.R) - int(c2.R)
	if diffR < 0 {
		diffR = -diffR
	}
	diffG := int(c1.G) - int(c2.G)
	if diffG < 0 {
		diffG = -diffG
	}
	diffB := int(c1.B) - int(c2.B)
	if diffB < 0 {
		diffB = -diffB
	}

	return diffR <= int(tolerance) && diffG <= int(tolerance) && diffB <= int(tolerance)
}

// modifyFlagColors applies noticeable color modifications to create incorrect version
func modifyFlagColors(img image.Image, correct bool) image.Image {
	if correct {
		return img
	}

	var allColors []color.RGBA = getDistinctColors(img)

	colorToBeModified := allColors[rand.Intn(len(allColors))]

	// Create a new image with modified colors
	bounds := img.Bounds()
	modified := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()

			origR := uint8(r >> 8)
			origG := uint8(g >> 8)
			origB := uint8(b >> 8)

			// Convert original color to RGBA for comparison
			origRGBA := color.RGBA{origR, origG, origB, uint8(a >> 8)}

			// Only modify pixels that match the selected color (with some tolerance for anti-aliasing)
			if !colorsAreSimilar(origRGBA, colorToBeModified, 20) {
				modified.Set(x, y, originalColor)
				continue
			}

			var newR, newG, newB uint8
			// red component
			if origR > 50 {
				newR = origR - 40 // darker
			} else if origR < 205 {
				newR = origR + 50 // brighter
			} else {
				newR = origR - 30
			}

			// green component
			if origG > 60 {
				newG = origG - 45 // reduce green significantly
			} else if origG < 195 {
				newG = origG + 60 // add green significantly
			} else {
				newG = origG - 35
			}

			// blue component
			if origB > 45 {
				newB = origB - 35 // reduce blue (warmer tone)
			} else if origB < 210 {
				newB = origB + 45 // Add blue (cooler tone)
			} else {
				newB = origB - 25
			}

			modifiedColor := color.RGBA{newR, newG, newB, uint8(a >> 8)}
			modified.Set(x, y, modifiedColor)
		}
	}

	return modified
}

// imageToBase64 converts an image to a base64 encoded string
func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
