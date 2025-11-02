package main

type Player struct {
	Name       string
	Correct    int
	Incorrect  int
	Total      int
	Percentage int
}

type GameState struct {
	Players       []Player // List of players
	CurrentPlayer int      // Index of current player
	GameStarted   bool     // Whether the game has started (players have been set up)
	TotalRounds   int      // Total number of rounds to play
	CurrentRound  int      // Current round number (1-indexed)
	GameOver      bool     // Whether the game has ended
	IsCorrect     bool     // Whether the current flag is the correct one
	CountryName   string   // Name of the current country
	FlagData      string   // Base64 encoded flag image being displayed
	OriginalFlag  string   // Base64 of original flag for comparison
	ModifiedFlag  string   // Base64 of modified flag for comparison
	ShowResult    bool     // Whether to show the result of the last guess
	ResultCorrect bool     // Whether the user's last guess was correct
	ResultMessage string   // Message to display about the result
}

type CountryFlag struct {
	Name    string // Country name
	FlagURL string // URL to the flag image
}
