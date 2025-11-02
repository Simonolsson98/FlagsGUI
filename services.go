package main

import (
	"fmt"
	"image"
	"math/rand"
	"strings"

	"github.com/pariz/gountries"
)

// CountryServiceImpl implements CountryService
type CountryServiceImpl struct {
	query *gountries.Query
}

// NewCountryService creates a new country service
func NewCountryService() CountryService {
	return &CountryServiceImpl{
		query: gountries.New(),
	}
}

// countryNameMappings handles special cases where flag URLs don't match country names
var countryNameMappings = map[string]string{
	"Taiwan":                           "Taiwan",
	"United States":                    "United_States_of_America",
	"United Kingdom":                   "United_Kingdom",
	"South Korea":                      "South_Korea",
	"North Korea":                      "North_Korea",
	"Czech Republic":                   "Czech_Republic",
	"Dominican Republic":               "Dominican_Republic",
	"Central African Republic":         "Central_African_Republic",
	"Democratic Republic of the Congo": "Democratic_Republic_of_the_Congo",
	"Costa Rica":                       "Costa_Rica",
	"Puerto Rico":                      "Puerto_Rico",
	"Hong Kong":                        "Hong_Kong",
	"Cape Verde":                       "Cape_Verde",
	"São Tomé and Príncipe":            "Sao_Tome_and_Principe",
	"Bosnia and Herzegovina":           "Bosnia_and_Herzegovina",
	"Trinidad and Tobago":              "Trinidad_and_Tobago",
	"Antigua and Barbuda":              "Antigua_and_Barbuda",
}

// getCleanCountryName returns the correct name for flag URL construction
func (s *CountryServiceImpl) getCleanCountryName(countryName string) string {
	// Check if we have a special mapping for this country
	if mappedName, exists := countryNameMappings[countryName]; exists {
		return mappedName
	}

	// Default cleaning for countries not in the special mapping
	cleanName := strings.ReplaceAll(countryName, " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "'", "")
	cleanName = strings.ReplaceAll(cleanName, ".", "")
	cleanName = strings.ReplaceAll(cleanName, "é", "e")
	cleanName = strings.ReplaceAll(cleanName, "ç", "c")
	cleanName = strings.ReplaceAll(cleanName, "ã", "a")
	cleanName = strings.ReplaceAll(cleanName, "õ", "o")

	return cleanName
}

// GetRandomCountry returns a random country and constructs its flag URL
func (s *CountryServiceImpl) GetRandomCountry() CountryFlag {
	allCountries := s.query.FindAllCountries()
	if len(allCountries) == 0 {
		// Fallback in case library fails
		return CountryFlag{"Sweden", "https://flagdownload.com/wp-content/uploads/Flag_of_Sweden-256x171.png"}
	}

	// Convert map to slice for random selection
	var countryList []gountries.Country
	for _, country := range allCountries {
		countryList = append(countryList, country)
	}

	// Pick a random country
	randomCountry := countryList[rand.Intn(len(countryList))]
	countryName := randomCountry.Name.Common

	// Get the clean name for URL construction
	cleanName := s.getCleanCountryName(countryName)

	// Construct the flag URL with 128px height as default (will fallback to other heights if needed)
	flagURL := fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s-256x128.png", cleanName)

	return CountryFlag{
		Name:    countryName,
		FlagURL: flagURL,
	}
}

// ImageServiceImpl implements ImageService
type ImageServiceImpl struct{}

// NewImageService creates a new image service
func NewImageService() ImageService {
	return &ImageServiceImpl{}
}

// DownloadFlag downloads a flag image with fallback for different dimensions
func (s *ImageServiceImpl) DownloadFlag(url string) (image.Image, error) {
	// Try the primary URL first
	img, err := downloadFlagImage(url)
	if err == nil {
		return img, nil
	}

	// If primary URL fails, try different height variations
	return s.downloadWithHeightFallbacks(url)
}

// downloadWithHeightFallbacks tries multiple height variations for 256px width flags
func (s *ImageServiceImpl) downloadWithHeightFallbacks(originalURL string) (image.Image, error) {
	if !strings.Contains(originalURL, "Flag_of_") || !strings.Contains(originalURL, "-256x") {
		return downloadFlagImage(originalURL)
	}

	// Extract the country name and base URL structure
	parts := strings.Split(originalURL, "Flag_of_")
	if len(parts) < 2 {
		return downloadFlagImage(originalURL)
	}

	countryPart := strings.Split(parts[1], "-256x")[0]
	baseURL := parts[0] + "Flag_of_" + countryPart + "-256x%d.png"
	for height := 130; height >= 126; height-- {
		url := fmt.Sprintf(baseURL, height)
		if img, err := downloadFlagImage(url); err == nil {
			return img, nil
		}
	}

	squareURL := parts[0] + "Flag_of_" + countryPart + "_Flat_Square-512x512.png"
	if img, err := downloadFlagImage(squareURL); err == nil {
		return img, nil
	}

	// Return the original error
	return downloadFlagImage(originalURL)
}

// ModifyColors applies color modifications to create incorrect version
func (s *ImageServiceImpl) ModifyColors(img image.Image, correct bool) image.Image {
	return modifyFlagColors(img, correct)
}

// ToBase64 converts an image to base64 encoded string
func (s *ImageServiceImpl) ToBase64(img image.Image) (string, error) {
	return imageToBase64(img)
}
