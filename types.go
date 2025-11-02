package main

type Player struct {
	Name       string
	Correct    int
	Incorrect  int
	Total      int
	Percentage int
}

type GameState struct {
	Players       []Player
	CurrentPlayer int
	GameStarted   bool
	TotalRounds   int
	CurrentRound  int
	GameOver      bool
	IsCorrect     bool
	CountryName   string
	FlagData      string
	OriginalFlag  string
	ModifiedFlag  string
	ShowResult    bool
	ResultCorrect bool
	ResultMessage string
}

type CountryFlag struct {
	Name    string
	FlagURL string
}
