package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/pariz/gountries"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Flag Quiz Game</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 900px;
            margin: 20px auto;
            padding: 20px;
            background-color: #f0f8ff;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 15px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            text-align: center;
            margin-bottom: 10px;
        }
        .subtitle {
            text-align: center;
            color: #7f8c8d;
            margin-bottom: 30px;
            font-style: italic;
        }
        .game-area {
            text-align: center;
            margin: 30px 0;
        }
        .flag-container {
            display: inline-block;
            border: 3px solid #34495e;
            border-radius: 10px;
            overflow: hidden;
            margin: 20px 0;
            box-shadow: 0 2px 10px rgba(0,0,0,0.2);
        }
        .flag-image {
            display: block;
            max-width: 100%;
            height: auto;
        }
        .question {
            font-size: 24px;
            color: #2c3e50;
            margin: 20px 0;
            font-weight: bold;
        }
        .country-name {
            font-size: 20px;
            color: #e74c3c;
            margin: 15px 0;
            font-weight: bold;
        }
        .buttons {
            margin: 30px 0;
        }
        .btn {
            font-size: 18px;
            padding: 15px 30px;
            margin: 10px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-weight: bold;
            transition: all 0.3s ease;
        }
        .btn-correct {
            background-color: #27ae60;
            color: white;
        }
        .btn-correct:hover {
            background-color: #229954;
        }
        .btn-incorrect {
            background-color: #e74c3c;
            color: white;
        }
        .btn-incorrect:hover {
            background-color: #c0392b;
        }
        .btn-new {
            background-color: #3498db;
            color: white;
        }
        .btn-new:hover {
            background-color: #2980b9;
        }
        .result {
            font-size: 20px;
            font-weight: bold;
            margin: 20px 0;
            padding: 15px;
            border-radius: 8px;
        }
        .result.correct {
            background-color: #d5f4e6;
            color: #27ae60;
            border: 2px solid #27ae60;
        }
        .result.incorrect {
            background-color: #fadbd8;
            color: #e74c3c;
            border: 2px solid #e74c3c;
        }
        .score {
            background-color: #ebf3fd;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            text-align: center;
        }
        .score h3 {
            margin: 0 0 10px 0;
            color: #2c3e50;
        }
        .stats {
            display: flex;
            justify-content: space-around;
            flex-wrap: wrap;
        }
        .stat {
            margin: 5px;
        }
        .stat-value {
            font-size: 24px;
            font-weight: bold;
            color: #3498db;
        }
        .stat-label {
            font-size: 14px;
            color: #7f8c8d;
        }
        .flag-comparison {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin: 20px 0;
            flex-wrap: wrap;
        }
        .flag-box {
            text-align: center;
            padding: 10px;
            border-radius: 8px;
            background-color: #f8f9fa;
            border: 2px solid #dee2e6;
            min-width: 120px;
        }
        .flag-box h4 {
            margin: 0 0 10px 0;
            font-size: 14px;
            color: #495057;
        }
        .flag-thumbnail {
            width: 80px;
            height: 80px;
            border: 2px solid #333;
            border-radius: 4px;
            object-fit: cover;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏴 Flag Quiz Game</h1>
        <div class="subtitle">Can you spot the fake flags?</div>
        
        <div class="score">
            <h3>Your Score</h3>
            <div class="stats">
                <div class="stat">
                    <div class="stat-value" id="correct">{{.Correct}}</div>
                    <div class="stat-label">Correct</div>
                </div>
                <div class="stat">
                    <div class="stat-value" id="incorrect">{{.Incorrect}}</div>
                    <div class="stat-label">Incorrect</div>
                </div>
                <div class="stat">
                    <div class="stat-value" id="total">{{.Total}}</div>
                    <div class="stat-label">Total</div>
                </div>
                <div class="stat">
                    <div class="stat-value" id="percentage">{{.Percentage}}%</div>
                    <div class="stat-label">Accuracy</div>
                </div>
            </div>
        </div>

        <div class="game-area">
            {{if .FlagData}}
                <div class="question">Is this the correct flag?</div>
                <div class="country-name">{{.CountryName}}</div>
                <div class="flag-container">
                    <img src="data:image/png;base64,{{.FlagData}}" alt="Flag" class="flag-image">
                </div>
                
                {{if .ShowResult}}
                    <div class="result {{if .ResultCorrect}}correct{{else}}incorrect{{end}}">
                        {{if .ResultCorrect}}
                            ✓ Correct! {{.ResultMessage}}
                        {{else}}
                            ✗ Wrong! {{.ResultMessage}}
                            {{if and .OriginalFlag .ModifiedFlag (not .IsCorrect)}}
                                <div class="flag-comparison">
                                    <div class="flag-box">
                                        <h4>Correct Flag</h4>
                                        <img src="data:image/png;base64,{{.OriginalFlag}}" alt="Correct Flag" class="flag-thumbnail">
                                    </div>
                                </div>
                            {{end}}
                        {{end}}
                    </div>
                    <button class="btn btn-new" onclick="location.href='/new'">Next Flag</button>
                {{else}}
                    <div class="buttons">
                        <button class="btn btn-correct" onclick="location.href='/guess?answer=correct'">Correct Flag</button>
                        <button class="btn btn-incorrect" onclick="location.href='/guess?answer=incorrect'">Fake Flag</button>
                    </div>
                {{end}}
            {{else}}
                <div class="question">Welcome to the Flag Quiz!</div>
                <p>Test your knowledge of world flags. Some flags will be correct, others will have subtle errors.</p>
                <button class="btn btn-new" onclick="location.href='/new'">Start Game</button>
            {{end}}
        </div>
    </div>
</body>
</html>
`

type GameState struct {
	IsCorrect     bool
	CountryName   string
	FlagData      string
	OriginalFlag  string // Base64 of original flag for comparison
	ModifiedFlag  string // Base64 of modified flag for comparison
	ShowResult    bool
	ResultCorrect bool
	ResultMessage string
	Correct       int
	Incorrect     int
	Total         int
	Percentage    int
}

var gameState = &GameState{}

type CountryFlag struct {
	Name    string
	FlagURL string
}

// Initialize the gountries query object
var query = gountries.New()

// Get a random country and construct its flag URL
func getRandomCountry() CountryFlag {
	allCountries := query.FindAllCountries()
	if len(allCountries) == 0 {
		// Fallback in case library fails
		return CountryFlag{"Sweden", "https://flagdownload.com/wp-content/uploads/Flag_of_Sweden_Flat_Square-256x256.png"}
	}

	// Convert map to slice for random selection
	var countryList []gountries.Country
	for _, country := range allCountries {
		countryList = append(countryList, country)
	}

	// Pick a random country
	randomCountry := countryList[rand.Intn(len(countryList))]
	countryName := randomCountry.Name.Common

	// Clean the country name for URL (replace spaces with underscores, handle special cases)
	cleanName := strings.ReplaceAll(countryName, " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "'", "")
	cleanName = strings.ReplaceAll(cleanName, ".", "")

	// Construct the flag URL
	flagURL := fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s_Flat_Square-256x256.png", cleanName)

	return CountryFlag{
		Name:    countryName,
		FlagURL: flagURL,
	}
}

// Download flag image from URL
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

// Get dominant colors from flag image, filtering out noise and anti-aliasing artifacts
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

// Apply noticeable color modifications to create incorrect version
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

func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, gameState)
}

func newGameHandler(w http.ResponseWriter, r *http.Request) {
	rand.NewSource(rand.NewSource(time.Now().UnixNano()).Int63())

	country := getRandomCountry()
	isCorrect := rand.Intn(2) == 0

	// Download the original flag
	originalImg, err := downloadFlagImage(country.FlagURL)
	for err != nil {
		log.Printf("Error downloading flag for %s: %v", country.Name, err)
		country := getRandomCountry()
		originalImg, err = downloadFlagImage(country.FlagURL)
	}

	// Create both original and modified versions for comparison
	originalFlagData, err := imageToBase64(originalImg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	modifiedImg := modifyFlagColors(originalImg, false) // Always create modified version
	modifiedFlagData, err := imageToBase64(modifiedImg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Choose which version to show
	var displayImg image.Image
	if isCorrect {
		displayImg = originalImg
	} else {
		displayImg = modifiedImg
	}

	flagData, err := imageToBase64(displayImg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gameState.IsCorrect = isCorrect
	gameState.CountryName = country.Name
	gameState.FlagData = flagData
	gameState.OriginalFlag = originalFlagData
	gameState.ModifiedFlag = modifiedFlagData
	gameState.ShowResult = false

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	answer := r.URL.Query().Get("answer")

	var userCorrect bool
	if answer == "correct" {
		userCorrect = gameState.IsCorrect
	} else {
		userCorrect = !gameState.IsCorrect
	}

	gameState.Total++
	if userCorrect {
		gameState.Correct++
		gameState.ResultCorrect = true
		if gameState.IsCorrect {
			gameState.ResultMessage = "This is indeed the correct " + gameState.CountryName + " flag!"
		} else {
			gameState.ResultMessage = "Good eye! This flag had incorrect colors."
		}
	} else {
		gameState.Incorrect++
		gameState.ResultCorrect = false
		if gameState.IsCorrect {
			gameState.ResultMessage = "This was actually the correct " + gameState.CountryName + " flag."
		} else {
			gameState.ResultMessage = "This flag had wrong colors - you missed it!"
		}
	}

	if gameState.Total > 0 {
		gameState.Percentage = (gameState.Correct * 100) / gameState.Total
	}

	gameState.ShowResult = true

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("Error opening browser: %v", err)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/new", newGameHandler)
	http.HandleFunc("/guess", guessHandler)

	port := ":8080"
	url := "http://localhost" + port

	fmt.Printf("Starting Flag Quiz Game server on %s\n", url)
	fmt.Println("Opening browser...")

	// Open browser after a short delay
	go func() {
		openBrowser(url)
	}()

	fmt.Printf("Server running on %s\n", url)
	fmt.Println("Press Ctrl+C to stop")

	log.Fatal(http.ListenAndServe(port, nil))
}
