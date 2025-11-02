package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/pariz/gountries"
)

var query = gountries.New()

// returns a random country and constructs its flag URL
func getRandomCountry() CountryFlag {
	allCountries := query.FindAllCountries()
	if len(allCountries) == 0 {
		// Fallback in case library fails
		return CountryFlag{"Sweden", "https://flagdownload.com/wp-content/uploads/Flag_of_Sweden_Flat_Square-256x256.png"}
	}

	var countryList []gountries.Country
	for _, country := range allCountries {
		countryList = append(countryList, country)
	}

	randomCountry := countryList[rand.Intn(len(countryList))]
	countryName := randomCountry.Name.Common

	cleanName := strings.ReplaceAll(countryName, " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "'", "")
	cleanName = strings.ReplaceAll(cleanName, ".", "")

	flagURL := fmt.Sprintf("https://flagdownload.com/wp-content/uploads/Flag_of_%s_Flat_Square-256x256.png", cleanName)

	return CountryFlag{
		Name:    countryName,
		FlagURL: flagURL,
	}
}
