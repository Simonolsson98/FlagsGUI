package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
)

func downloadFlagImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

type ColorInfo struct {
	Color color.RGBA
	Count int
}

func getDistinctColors(img image.Image) []color.RGBA {
	bounds := img.Bounds()
	colorMap := make(map[[3]uint8]int)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			if uint8(a>>8) < 128 {
				continue
			}

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
	threshold := totalPixels / 200
	if threshold < 10 {
		threshold = 10
	}

	var colorInfos []ColorInfo
	for rgb, count := range colorMap {
		if count > threshold {
			colorInfos = append(colorInfos, ColorInfo{
				Color: color.RGBA{
					R: rgb[0],
					G: rgb[1],
					B: rgb[2],
					A: 255,
				},
				Count: count,
			})
		}
	}

	// Sort by pixel count (most prominent first)
	for i := 0; i < len(colorInfos); i++ {
		for j := i + 1; j < len(colorInfos); j++ {
			if colorInfos[j].Count > colorInfos[i].Count {
				colorInfos[i], colorInfos[j] = colorInfos[j], colorInfos[i]
			}
		}
	}

	var colors []color.RGBA
	for _, info := range colorInfos {
		colors = append(colors, info.Color)
	}

	return colors
}

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

func rgbToHSV(r, g, b uint8) (h, s, v float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := rf
	if gf > max {
		max = gf
	}
	if bf > max {
		max = bf
	}

	min := rf
	if gf < min {
		min = gf
	}
	if bf < min {
		min = bf
	}

	v = max
	if max == 0 {
		s = 0
	} else {
		s = (max - min) / max
	}

	if max == min {
		h = 0
	} else if max == rf {
		h = 60 * ((gf - bf) / (max - min))
	} else if max == gf {
		h = 60 * (2 + (bf-rf)/(max-min))
	} else {
		h = 60 * (4 + (rf-gf)/(max-min))
	}

	if h < 0 {
		h += 360
	}

	return h, s, v
}

func hsvToRGB(h, s, v float64) (r, g, b uint8) {
	c := v * s
	x := c * (1 - abs(mod(h/60, 2)-1))
	m := v - c

	var rf, gf, bf float64

	if h >= 0 && h < 60 {
		rf, gf, bf = c, x, 0
	} else if h >= 60 && h < 120 {
		rf, gf, bf = x, c, 0
	} else if h >= 120 && h < 180 {
		rf, gf, bf = 0, c, x
	} else if h >= 180 && h < 240 {
		rf, gf, bf = 0, x, c
	} else if h >= 240 && h < 300 {
		rf, gf, bf = x, 0, c
	} else {
		rf, gf, bf = c, 0, x
	}

	r = uint8((rf + m) * 255)
	g = uint8((gf + m) * 255)
	b = uint8((bf + m) * 255)

	return r, g, b
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func mod(x, y float64) float64 {
	return x - y*float64(int(x/y))
}

func drasticColorChange(originalColor color.RGBA) color.RGBA {
	contrastColors := []color.RGBA{
		{255, 0, 0, 255},     // Red
		{0, 255, 0, 255},     // Green
		{0, 0, 255, 255},     // Blue
		{255, 255, 0, 255},   // Yellow
		{255, 0, 255, 255},   // Magenta
		{0, 255, 255, 255},   // Cyan
		{255, 165, 0, 255},   // Orange
		{128, 0, 128, 255},   // Purple
		{255, 192, 203, 255}, // Pink
		{0, 128, 0, 255},     // Dark Green
		{139, 69, 19, 255},   // Brown
		{255, 20, 147, 255},  // Deep Pink
	}
	for _, candidate := range contrastColors {
		if !colorsAreSimilar(originalColor, candidate, 100) {
			return candidate
		}
	}

	return color.RGBA{255 - originalColor.R, 255 - originalColor.G, 255 - originalColor.B, originalColor.A}
}

func adjustColorShade(originalColor color.RGBA) color.RGBA {
	h, s, v := rgbToHSV(originalColor.R, originalColor.G, originalColor.B)

	if s < 0.1 {
		s = 0.4 + rand.Float64()*0.4
		h = float64(rand.Intn(360))
	}

	adjustBrightness := rand.Float64() < 0.5

	if adjustBrightness {
		if rand.Float64() < 0.5 {
			v = v * (0.05 + rand.Float64()*0.15)
		} else {
			v = min(1.0, v+0.4)
			s = s * 0.5
		}
	} else {
		if rand.Float64() < 0.5 {
			s = 0.05 + rand.Float64()*0.45
		} else {
			if v >= 0.8 {
				v = v * (0.2 + rand.Float64()*0.3)
			} else if s >= 0.7 {
				v = min(1.0, v+0.5)
			} else {
				s = min(1.0, s*(1.8+rand.Float64()*0.7))
			}
		}
	}

	r, g, b := hsvToRGB(h, s, v)
	return color.RGBA{r, g, b, originalColor.A}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func modifyFlagColors(img image.Image, correct bool) image.Image {
	if correct {
		return img
	}

	var allColors []color.RGBA = getDistinctColors(img)
	if len(allColors) == 0 {
		log.Printf("⚠️  No distinct colors found in image")
		return img
	}

	log.Printf("🎨 Found %d distinct colors to potentially modify", len(allColors))

	var suitableColors []color.RGBA
	for _, c := range allColors {
		brightness := (int(c.R) + int(c.G) + int(c.B)) / 3
		if brightness > 20 && brightness < 240 {
			suitableColors = append(suitableColors, c)
		}
	}

	var colorToBeModified color.RGBA
	if len(suitableColors) > 0 {
		randomIndex := rand.Intn(len(suitableColors))
		colorToBeModified = suitableColors[randomIndex]
		log.Printf("🎲 Randomly selected color %d out of %d suitable colors", randomIndex+1, len(suitableColors))
	}

	if colorToBeModified.R == 0 && colorToBeModified.G == 0 && colorToBeModified.B == 0 && len(allColors) > 0 {
		colorToBeModified = allColors[0]
		log.Printf("🎲 Fallback - using most prominent color")
	}

	log.Printf("🎯 Selected color to modify: R=%d, G=%d, B=%d",
		colorToBeModified.R, colorToBeModified.G, colorToBeModified.B)

	useDrasticChange := rand.Float64() < 0.5

	var newColor color.RGBA
	if useDrasticChange {
		newColor = drasticColorChange(colorToBeModified)
		log.Printf("💥 Applied DRASTIC change: R=%d, G=%d, B=%d -> R=%d, G=%d, B=%d",
			colorToBeModified.R, colorToBeModified.G, colorToBeModified.B,
			newColor.R, newColor.G, newColor.B)
	} else {
		newColor = adjustColorShade(colorToBeModified)
		log.Printf("🎨 Applied SHADE ADJUSTMENT: R=%d, G=%d, B=%d -> R=%d, G=%d, B=%d",
			colorToBeModified.R, colorToBeModified.G, colorToBeModified.B,
			newColor.R, newColor.G, newColor.B)
	}

	bounds := img.Bounds()
	modified := image.NewRGBA(bounds)
	modifiedPixels := 0
	totalPixels := (bounds.Max.X - bounds.Min.X) * (bounds.Max.Y - bounds.Min.Y)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()

			origR := uint8(r >> 8)
			origG := uint8(g >> 8)
			origB := uint8(b >> 8)

			// Convert original color to RGBA for comparison
			origRGBA := color.RGBA{origR, origG, origB, uint8(a >> 8)}

			// Check if this pixel belongs to the same color family as our target
			// Since target colors are quantized, we need to check if this pixel
			// would quantize to the same value
			quantizedR := (origR / 16) * 16
			quantizedG := (origG / 16) * 16
			quantizedB := (origB / 16) * 16
			quantizedPixel := color.RGBA{quantizedR, quantizedG, quantizedB, 255}

			isMatch := colorsAreSimilar(quantizedPixel, colorToBeModified, 32) ||
				colorsAreSimilar(origRGBA, colorToBeModified, 100)

			if !isMatch {
				modified.Set(x, y, originalColor)
				continue
			}

			modified.Set(x, y, newColor)
			modifiedPixels++
		}
	}

	log.Printf("✏️  Modified %d out of %d pixels (%.2f%%)",
		modifiedPixels, totalPixels, float64(modifiedPixels)/float64(totalPixels)*100)

	if modifiedPixels == 0 {
		log.Printf("⚠️  WARNING: No pixels were modified!")
	}

	return modified
}

func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
