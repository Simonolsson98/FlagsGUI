package main

import (
	"fmt"
	"testing"
)

// TestDownloadFlag tests the flag downloading with height fallbacks
func TestDownloadFlag(t *testing.T) {
	fmt.Println("Testing Flag URL Resolution with Height Fallbacks...")
	fmt.Println("============================================================")

	// Create services
	imageService := NewImageService()

	// Test a few countries
	testCountries := []string{"Germany", "Switzerland", "Nepal", "Vatican City"}

	for _, testCountryName := range testCountries {
		fmt.Printf("\n🏁 Testing %s:\n", testCountryName)

		// Try different height combinations for this country
		baseURL := fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s-256x", testCountryName)

		fmt.Printf("Trying: %s\n", baseURL)
		_, err := imageService.DownloadFlag(baseURL)
		if err == nil {
			fmt.Printf("✅ Success with URL: %s\n", baseURL)
		} else {
			t.Error(fmt.Errorf("❌ Failed: %v\n", err))
		}
	}
}
