package main

import (
	"fmt"
	"html/template"
	"image"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
		var tmpl *template.Template
		var err error

		if !deps.GameState.GameStarted {
			// Show player setup page
			tmpl, err = template.New("setup").Parse(setupTemplate)
		} else {
			// Show game page
			tmpl, err = template.New("index").Parse(htmlTemplate)
		}

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

		// Move to next player if this is not the first game
		if deps.GameState.ShowResult && len(deps.GameState.Players) > 0 {
			deps.GameState.CurrentPlayer = (deps.GameState.CurrentPlayer + 1) % len(deps.GameState.Players)

			// Check if we've completed a full round (back to first player)
			if deps.GameState.CurrentPlayer == 0 {
				deps.GameState.CurrentRound++

				// Check if game is over
				if deps.GameState.CurrentRound > deps.GameState.TotalRounds {
					deps.GameState.GameOver = true
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			}
		}

		var country CountryFlag
		if debugCountry != "" {
			// Debug mode: create country with specified name
			// Use the same URL format as the country service
			cleanName := strings.ReplaceAll(debugCountry, " ", "_")
			country = CountryFlag{
				Name:    debugCountry,
				FlagURL: fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s-256x128.png", cleanName),
			}
			log.Printf("🐛 DEBUG: Using country '%s' with URL: %s", country.Name, country.FlagURL)
		} else {
			country = deps.CountryService.GetRandomCountry()
		}

		var isCorrect bool
		if debugCountry != "" {
			isCorrect = false // In debug mode, always show the modified flag to test color changes
			log.Printf("🐛 DEBUG: Forcing modified flag display for testing")
		} else {
			isCorrect = rand.Intn(2) == 0
		}

		// Download the original flag
		originalImg, err := deps.ImageService.DownloadFlag(country.FlagURL)
		for err != nil {
			log.Printf("Error downloading flag for %s: %v", country.Name, err)
			if debugCountry != "" {
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

// setupPlayersHandler handles player setup
func setupPlayersHandler(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Reset game state and show setup page
			deps.GameState.GameStarted = false
			deps.GameState.GameOver = false
			deps.GameState.Players = nil
			deps.GameState.CurrentPlayer = 0
			deps.GameState.CurrentRound = 0
			deps.GameState.TotalRounds = 0
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r.ParseForm()
		playerNames := r.Form["playerName"]

		// Get number of rounds
		totalRounds := 10 // Default
		if roundsStr := r.FormValue("numRounds"); roundsStr != "" {
			if rounds, err := strconv.Atoi(roundsStr); err == nil && rounds > 0 {
				totalRounds = rounds
			}
		}

		// Limit to 4 players and filter empty names
		var players []Player
		for i, name := range playerNames {
			if i >= 4 {
				break
			}
			name = strings.TrimSpace(name)
			if name != "" {
				players = append(players, Player{Name: name})
			}
		}

		if len(players) == 0 {
			// No valid players, stay on setup page
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		deps.GameState.Players = players
		deps.GameState.CurrentPlayer = 0
		deps.GameState.GameStarted = true
		deps.GameState.TotalRounds = totalRounds
		deps.GameState.CurrentRound = 1
		deps.GameState.GameOver = false
		deps.GameState.ShowResult = false

		http.Redirect(w, r, "/new", http.StatusSeeOther)
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

		// Update current player's score
		player := &deps.GameState.Players[deps.GameState.CurrentPlayer]
		player.Total++
		if userCorrect {
			player.Correct++
			deps.GameState.ResultCorrect = true
			if deps.GameState.IsCorrect {
				deps.GameState.ResultMessage = player.Name + ": This is indeed the correct " + deps.GameState.CountryName + " flag!"
			} else {
				deps.GameState.ResultMessage = player.Name + ": Good eye! This flag had incorrect colors."
			}
		} else {
			player.Incorrect++
			deps.GameState.ResultCorrect = false
			if deps.GameState.IsCorrect {
				deps.GameState.ResultMessage = player.Name + ": This was actually the correct " + deps.GameState.CountryName + " flag."
			} else {
				deps.GameState.ResultMessage = player.Name + ": This flag had wrong colors - you missed it!"
			}
		}

		if player.Total > 0 {
			player.Percentage = (player.Correct * 100) / player.Total
		}

		deps.GameState.ShowResult = true

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
