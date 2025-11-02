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

type Dependencies struct {
	GameState      *GameState
	CountryService CountryService
	ImageService   ImageService
}

type CountryService interface {
	GetRandomCountry() CountryFlag
}

type ImageService interface {
	DownloadFlag(url string) (image.Image, error)
	ModifyColors(img image.Image, correct bool) image.Image
	ToBase64(img image.Image) (string, error)
}

func indexHandler(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tmpl *template.Template
		var err error

		if !deps.GameState.GameStarted {
			tmpl, err = template.New("setup").Parse(setupTemplate)
		} else {
			tmpl, err = template.New("index").Parse(htmlTemplate)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, deps.GameState)
	}
}

func newGameHandler(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rand.NewSource(rand.NewSource(time.Now().UnixNano()).Int63())

		if !handlePlayerRotation(deps, w, r) {
			return
		}

		country := getCountry(deps)

		originalImg, actualCountry, err := downloadFlagWithRetry(deps, country)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		deps.GameState.IsCorrect = shouldShowCorrectFlag()
		deps.GameState.CountryName = actualCountry.Name

		if err := prepareFlagData(deps, originalImg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		deps.GameState.ShowResult = false

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func handlePlayerRotation(deps *Dependencies, w http.ResponseWriter, r *http.Request) bool {
	if !deps.GameState.ShowResult || len(deps.GameState.Players) == 0 {
		return true
	}

	deps.GameState.CurrentPlayer = (deps.GameState.CurrentPlayer + 1) % len(deps.GameState.Players)

	// full round completed
	if deps.GameState.CurrentPlayer == 0 {
		deps.GameState.CurrentRound++

		// game over
		if deps.GameState.CurrentRound > deps.GameState.TotalRounds {
			deps.GameState.GameOver = true
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return false
		}
	}

	return true
}

func getCountry(deps *Dependencies) CountryFlag {
	if debugCountry != "" {
		cleanName := strings.ReplaceAll(debugCountry, " ", "_")
		country := CountryFlag{
			Name:    debugCountry,
			FlagURL: fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s-512x256.png", cleanName),
		}
		log.Printf("🐛 DEBUG: Using country '%s' with URL: %s", country.Name, country.FlagURL)
		return country
	}
	return deps.CountryService.GetRandomCountry()
}

func shouldShowCorrectFlag() bool {
	if debugCountry != "" {
		log.Printf("🐛 DEBUG: Forcing modified flag display for testing")
		return false
	}
	return rand.Intn(2) == 0
}

func downloadFlagWithRetry(deps *Dependencies, country CountryFlag) (image.Image, CountryFlag, error) {
	originalImg, err := deps.ImageService.DownloadFlag(country.FlagURL)
	for err != nil {
		log.Printf("Error downloading flag for %s: %v", country.Name, err)
		if debugCountry != "" {
			return nil, country, fmt.Errorf("failed to download flag for debug country %s", debugCountry)
		}
		country = deps.CountryService.GetRandomCountry()
		originalImg, err = deps.ImageService.DownloadFlag(country.FlagURL)
	}
	return originalImg, country, nil
}

func prepareFlagData(deps *Dependencies, originalImg image.Image) error {
	originalFlagData, err := deps.ImageService.ToBase64(originalImg)
	if err != nil {
		return err
	}

	modifiedImg := deps.ImageService.ModifyColors(originalImg, false)
	modifiedFlagData, err := deps.ImageService.ToBase64(modifiedImg)
	if err != nil {
		return err
	}

	var displayImg image.Image
	if deps.GameState.IsCorrect {
		displayImg = originalImg
	} else {
		displayImg = modifiedImg
	}

	flagData, err := deps.ImageService.ToBase64(displayImg)
	if err != nil {
		return err
	}

	deps.GameState.FlagData = flagData
	deps.GameState.OriginalFlag = originalFlagData
	deps.GameState.ModifiedFlag = modifiedFlagData

	return nil
}

func setupPlayersHandler(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			resetGameState(deps.GameState)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r.ParseForm()
		players := parsePlayerNames(r.Form["playerName"])
		if len(players) == 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		totalRounds := parseRoundsCount(r.FormValue("numRounds"))
		initializeGameState(deps.GameState, players, totalRounds)

		http.Redirect(w, r, "/new", http.StatusSeeOther)
	}
}

func resetGameState(state *GameState) {
	state.GameStarted = false
	state.GameOver = false
	state.Players = nil
	state.CurrentPlayer = 0
	state.CurrentRound = 0
	state.TotalRounds = 0
}

func parsePlayerNames(names []string) []Player {
	var players []Player
	for i, name := range names {
		if i >= 4 {
			break
		}
		name = strings.TrimSpace(name)
		if name != "" {
			players = append(players, Player{Name: name})
		}
	}
	return players
}

func parseRoundsCount(roundsStr string) int {
	if roundsStr != "" {
		if rounds, err := strconv.Atoi(roundsStr); err == nil && rounds > 0 {
			return rounds
		}
	}
	return 10
}

func initializeGameState(state *GameState, players []Player, totalRounds int) {
	state.Players = players
	state.CurrentPlayer = 0
	state.GameStarted = true
	state.TotalRounds = totalRounds
	state.CurrentRound = 1
	state.GameOver = false
	state.ShowResult = false
}

func guessHandler(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		answer := r.URL.Query().Get("answer")
		userCorrect := evaluateGuess(answer, deps.GameState.IsCorrect)

		player := &deps.GameState.Players[deps.GameState.CurrentPlayer]
		updatePlayerScore(player, userCorrect)

		deps.GameState.ResultCorrect = userCorrect
		deps.GameState.ResultMessage = generateResultMessage(player.Name, userCorrect, deps.GameState.IsCorrect, deps.GameState.CountryName)
		deps.GameState.ShowResult = true

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func evaluateGuess(answer string, isCorrectFlag bool) bool {
	if answer == "correct" {
		return isCorrectFlag
	}
	return !isCorrectFlag
}

func updatePlayerScore(player *Player, correct bool) {
	player.Total++
	if correct {
		player.Correct++
	} else {
		player.Incorrect++
	}

	if player.Total > 0 {
		player.Percentage = (player.Correct * 100) / player.Total
	}
}

func generateResultMessage(playerName string, userCorrect, flagCorrect bool, countryName string) string {
	if userCorrect {
		if flagCorrect {
			return fmt.Sprintf("%s: This is indeed the correct %s flag!", playerName, countryName)
		}
		return fmt.Sprintf("%s: Good eye! This flag had incorrect colors.", playerName)
	}

	if flagCorrect {
		return fmt.Sprintf("%s: This was actually the correct %s flag.", playerName, countryName)
	}
	return fmt.Sprintf("%s: This flag had wrong colors - you missed it!", playerName)
}
