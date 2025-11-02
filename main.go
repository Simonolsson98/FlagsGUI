package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Create dependencies
	gameState := &GameState{}
	countryService := NewCountryService()
	imageService := NewImageService()

	deps := &Dependencies{
		GameState:      gameState,
		CountryService: countryService,
		ImageService:   imageService,
	}

	// Register handlers with dependency injection
	http.HandleFunc("/", indexHandler(deps))
	http.HandleFunc("/new", newGameHandler(deps))
	http.HandleFunc("/guess", guessHandler(deps))

	port := ":8080"
	url := "http://localhost" + port

	fmt.Printf("Starting Flag Quiz Game server (with DI) on %s\n", url)
	fmt.Println("Opening browser...")

	// Open browser after a short delay
	go func() {
		openBrowser(url)
	}()

	fmt.Printf("Server running on %s\n", url)
	fmt.Println("Press Ctrl+C to stop")

	log.Fatal(http.ListenAndServe(port, nil))
}
