package main

type GameState struct {
	IsCorrect     bool   // Whether the current flag is the correct one
	CountryName   string // Name of the current country
	FlagData      string // Base64 encoded flag image being displayed
	OriginalFlag  string // Base64 of original flag for comparison
	ModifiedFlag  string // Base64 of modified flag for comparison
	ShowResult    bool   // Whether to show the result of the last guess
	ResultCorrect bool   // Whether the user's last guess was correct
	ResultMessage string // Message to display about the result
	Correct       int    // Number of correct guesses
	Incorrect     int    // Number of incorrect guesses
	Total         int    // Total number of guesses
	Percentage    int    // Accuracy percentage
}

type CountryFlag struct {
	Name    string // Country name
	FlagURL string // URL to the flag image
}
