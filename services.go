package main

import (
	"fmt"
	"image"
	"math/rand"
	"strings"

	"github.com/pariz/gountries"
)

type CountryServiceImpl struct {
	query *gountries.Query
}

func NewCountryService() CountryService {
	return &CountryServiceImpl{
		query: gountries.New(),
	}
}

var countryNameMappings = map[string]string{
	"Taiwan":                   "Taiwan",
	"United States":            "United_States_of_America",
	"United Kingdom":           "United_Kingdom",
	"South Korea":              "South_Korea",
	"North Korea":              "North_Korea",
	"Czech Republic":           "Czech_Republic",
	"Dominican Republic":       "Dominican_Republic",
	"Central African Republic": "Central_African_Republic",
	"DR Congo":                 "Democratic_Republic_of_the_Congo",
	"Costa Rica":               "Costa_Rica",
	"Puerto Rico":              "Puerto_Rico",
	"Hong Kong":                "Hong_Kong",
	"Cape Verde":               "Cape_Verde",
	"São Tomé and Príncipe":    "Sao_Tome_and_Principe",
	"Bosnia and Herzegovina":   "Bosnia_and_Herzegovina",
	"Trinidad and Tobago":      "Trinidad_and_Tobago",
	"Antigua and Barbuda":      "Antigua_and_Barbuda",
}

func (s *CountryServiceImpl) getCleanCountryName(countryName string) string {
	if mappedName, exists := countryNameMappings[countryName]; exists {
		return mappedName
	}

	cleanName := strings.ReplaceAll(countryName, " ", "_")	FlagURL: fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s-256x128.png", cleanName),
	cleanName = strings.ReplaceAll(cleanName, "'", "")
	cleanName = strings.ReplaceAll(cleanName, ".", "")
	cleanName = strings.ReplaceAll(cleanName, "é", "e")
	cleanName = strings.ReplaceAll(cleanName, "ç", "c")
	cleanName = strings.ReplaceAll(cleanName, "ã", "a")
	cleanName = strings.ReplaceAll(cleanName, "õ", "o")

	return cleanName
}

func (s *CountryServiceImpl) GetRandomCountry() CountryFlag {
	allCountries := s.query.FindAllCountries()
	if len(allCountries) == 0 {
		return CountryFlag{"Sweden", "https://flagdownload.com/wp-content/uploads/Flag_of_Sweden-256x171.png"}
	}

	var countryList []gountries.Country
	for _, country := range allCountries {
		countryList = append(countryList, country)
	}

	randomCountry := countryList[rand.Intn(len(countryList))]
	countryName := randomCountry.Name.Common
	cleanName := s.getCleanCountryName(countryName)
	flagURL := fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s-512x256.png", cleanName)

	return CountryFlag{
		Name:    countryName,
		FlagURL: flagURL,
	}
}

type ImageServiceImpl struct{}

func NewImageService() ImageService {
	return &ImageServiceImpl{}
}

func (s *ImageServiceImpl) DownloadFlag(url string) (image.Image, error) {
	img, err := downloadFlagImage(url)
	if err == nil {
		return img, nil
	}
	return s.downloadWithHeightFallbacks(url)
}

func (s *ImageServiceImpl) downloadWithHeightFallbacks(originalURL string) (image.Image, error) {
	if !strings.Contains(originalURL, "Flag_of_") || !strings.Contains(originalURL, "-512x") {
		return downloadFlagImage(originalURL)
	}

	parts := strings.Split(originalURL, "Flag_of_")
	if len(parts) < 2 {
		return downloadFlagImage(originalURL)
	}

	countryPart := strings.Split(parts[1], "-512x")[0]
	baseURL := parts[0] + "Flag_of_" + countryPart + "-512x%d.png"
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

	fmt.Printf("country flag: %s was not found\n", countryPart)
	return downloadFlagImage(originalURL)
}

func (s *ImageServiceImpl) ModifyColors(img image.Image, correct bool) image.Image {
	return modifyFlagColors(img, correct)
}

func (s *ImageServiceImpl) ToBase64(img image.Image) (string, error) {
	return imageToBase64(img)
}
