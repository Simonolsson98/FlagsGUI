package main

import (
	"fmt"
	"html/template"
	"image"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Dependencies struct holds all the dependencies our handlers need
type Dependencies struct {
	GameState      *GameState
	CountryService CountryService
	ImageService   ImageService
}

// CountryService interface defines country-related operations
type CountryService interface {
	GetRandomCountry() CountryFlag
}

// ImageService interface defines image-related operations
type ImageService interface {
	DownloadFlag(url string) (image.Image, error)
	ModifyColors(img image.Image, correct bool) image.Image
	ToBase64(img image.Image) (string, error)
}

// indexHandler returns a handler function with injected dependencies
func indexHandler(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("index").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, deps.GameState)
	}
}

// newGameHandler returns a handler function with injected dependencies
func newGameHandler(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rand.NewSource(rand.NewSource(time.Now().UnixNano()).Int63())

		var country CountryFlag
		if debugCountry != "" {
			// Debug mode: create country with specified name
			country = CountryFlag{
				Name:    debugCountry,
				FlagURL: fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s-256x128.png", debugCountry),
			}
			log.Printf("🐛 DEBUG: Using country '%s'", country.Name)
		} else {
			country = deps.CountryService.GetRandomCountry()
		}

		isCorrect := rand.Intn(2) == 0

		// Download the original flag
		originalImg, err := deps.ImageService.DownloadFlag(country.FlagURL)
		for err != nil {
			log.Printf("Error downloading flag for %s: %v", country.Name, err)
			if debugCountry != "" {
				// In debug mode, don't fall back to random country
				http.Error(w, fmt.Sprintf("Failed to download flag for debug country %s", debugCountry), http.StatusInternalServerError)
				return
			}
			country = deps.CountryService.GetRandomCountry()
			originalImg, err = deps.ImageService.DownloadFlag(country.FlagURL)
		}

		// Create both original and modified versions for comparison
		originalFlagData, err := deps.ImageService.ToBase64(originalImg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		modifiedImg := deps.ImageService.ModifyColors(originalImg, false)
		modifiedFlagData, err := deps.ImageService.ToBase64(modifiedImg)
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

		flagData, err := deps.ImageService.ToBase64(displayImg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		deps.GameState.IsCorrect = isCorrect
		deps.GameState.CountryName = country.Name
		deps.GameState.FlagData = flagData
		deps.GameState.OriginalFlag = originalFlagData
		deps.GameState.ModifiedFlag = modifiedFlagData
		deps.GameState.ShowResult = false

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// guessHandler returns a handler function with injected dependencies
func guessHandler(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		answer := r.URL.Query().Get("answer")

		var userCorrect bool
		if answer == "correct" {
			userCorrect = deps.GameState.IsCorrect
		} else {
			userCorrect = !deps.GameState.IsCorrect
		}

		deps.GameState.Total++
		if userCorrect {
			deps.GameState.Correct++
			deps.GameState.ResultCorrect = true
			if deps.GameState.IsCorrect {
				deps.GameState.ResultMessage = "This is indeed the correct " + deps.GameState.CountryName + " flag!"
			} else {
				deps.GameState.ResultMessage = "Good eye! This flag had incorrect colors."
			}
		} else {
			deps.GameState.Incorrect++
			deps.GameState.ResultCorrect = false
			if deps.GameState.IsCorrect {
				deps.GameState.ResultMessage = "This was actually the correct " + deps.GameState.CountryName + " flag."
			} else {
				deps.GameState.ResultMessage = "This flag had wrong colors - you missed it!"
			}
		}

		if deps.GameState.Total > 0 {
			deps.GameState.Percentage = (deps.GameState.Correct * 100) / deps.GameState.Total
		}

		deps.GameState.ShowResult = true

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
