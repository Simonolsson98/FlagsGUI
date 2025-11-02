package main

import (
	"html/template"
	"image"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// indexHandler serves the main page
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, gameState)
}

// newGameHandler generates a new flag for the quiz
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

// guessHandler processes the user's guess
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
