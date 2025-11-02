package main

import (
	"fmt"
	"log"
	"net/http"
)

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
