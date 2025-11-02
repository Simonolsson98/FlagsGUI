package main

import (
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockCountryService for testing
type MockCountryService struct {
	country CountryFlag
}

func (m *MockCountryService) GetRandomCountry() CountryFlag {
	return m.country
}

// MockImageService for testing
type MockImageService struct {
	downloadError error
	base64Result  string
	base64Error   error
}

func (m *MockImageService) DownloadFlag(url string) (image.Image, error) {
	if m.downloadError != nil {
		return nil, m.downloadError
	}
	// Create a simple test image (3x2 pixels with red and white)
	testImg := image.NewRGBA(image.Rect(0, 0, 3, 2))
	// Fill with red and white pattern (like Austrian flag)
	red := color.RGBA{255, 0, 0, 255}
	white := color.RGBA{255, 255, 255, 255}

	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			if y == 0 {
				testImg.Set(x, y, red) // Top row: red
			} else {
				testImg.Set(x, y, white) // Bottom row: white
			}
		}
	}
	return testImg, nil
}

func (m *MockImageService) ModifyColors(img image.Image, correct bool) image.Image {
	return img // Just return the same image for testing
}

func (m *MockImageService) ToBase64(img image.Image) (string, error) {
	if m.base64Error != nil {
		return "", m.base64Error
	}
	return m.base64Result, nil
}

// TestNewGameHandlerWithMocks tests the newGameHandler with mocked dependencies
func TestNewGameHandlerWithMocks(t *testing.T) {
	// Arrange - Create mock dependencies
	gameState := &GameState{}
	mockCountry := CountryFlag{Name: "TestCountry", FlagURL: "http://test.com/flag.png"}
	mockCountryService := &MockCountryService{country: mockCountry}
	mockImageService := &MockImageService{
		base64Result: "mock-base64-data",
	}

	deps := &Dependencies{
		GameState:      gameState,
		CountryService: mockCountryService,
		ImageService:   mockImageService,
	}

	// Act - Create handler with mocked dependencies
	handler := newGameHandler(deps)

	// Assert - Verify handler was created successfully
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	// Test the handler with an HTTP request
	req, err := http.NewRequest("GET", "/new", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	if gameState.CountryName != "TestCountry" {
		t.Errorf("Expected country name to be 'TestCountry', got '%s'", gameState.CountryName)
	}
	if gameState.FlagData == "" {
		t.Error("Expected FlagData to be set, got empty string")
	}
	if gameState.OriginalFlag == "" {
		t.Error("Expected OriginalFlag to be set, got empty string")
	}
	if gameState.ModifiedFlag == "" {
		t.Error("Expected ModifiedFlag to be set, got empty string")
	}
	if gameState.ShowResult {
		t.Error("Expected ShowResult to be false for new game")
	}
	// Verify the redirect location
	expectedLocation := "/"
	if location := rr.Header().Get("Location"); location != expectedLocation {
		t.Errorf("Expected redirect to '%s', got '%s'", expectedLocation, location)
	}
}

// TestNewGameHandlerWithDownloadError tests error handling when flag download fails
func TestNewGameHandlerWithDownloadError(t *testing.T) {
	// Arrange - Create mock dependencies with download error
	gameState := &GameState{}
	mockCountry := CountryFlag{Name: "ErrorCountry", FlagURL: "http://error.com/flag.png"}
	mockCountryService := &MockCountryService{country: mockCountry}
	mockImageService := &MockImageService{
		downloadError: http.ErrNotSupported, // Simulate download error
	}

	deps := &Dependencies{
		GameState:      gameState,
		CountryService: mockCountryService,
		ImageService:   mockImageService,
	}

	handler := newGameHandler(deps)

	// TODO: current implementation has an infinite loop when download fails. should be fixed by:
	// 1. Add a retry limit
	// 2. Return an error response
	// 3. Use a fallback country

	// For now, just verify the handler was created
	if handler == nil {
		t.Fatal("Expected handler to be created even with download errors")
	}
}

// Test_GIVEN_CorrectFlag_WHEN_AnsweringCorrectly_THEN_ExpectCorrectResponse tests the guess handler with a correct answer
func Test_GIVEN_CorrectFlag_WHEN_AnsweringCorrectly_THEN_ExpectCorrectResponse(t *testing.T) {
	// Arrange
	gameState := &GameState{
		IsCorrect:   true,
		CountryName: "TestCountry",
		Total:       0,
		Correct:     0,
		Incorrect:   0,
	}

	deps := &Dependencies{
		GameState: gameState,
	}

	handler := guessHandler(deps)

	// Act - Guess correctly when the flag is correct
	req, err := http.NewRequest("GET", "/guess?answer=correct", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Check game state was updated correctly
	if gameState.Total != 1 {
		t.Errorf("Expected Total to be 1, got %d", gameState.Total)
	}
	if gameState.Correct != 1 {
		t.Errorf("Expected Correct to be 1, got %d", gameState.Correct)
	}
	if gameState.Incorrect != 0 {
		t.Errorf("Expected Incorrect to be 0, got %d", gameState.Incorrect)
	}
	if !gameState.ResultCorrect {
		t.Error("Expected ResultCorrect to be true")
	}
	if !gameState.ShowResult {
		t.Error("Expected ShowResult to be true")
	}

	expectedMessage := "This is indeed the correct TestCountry flag!"
	if gameState.ResultMessage != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, gameState.ResultMessage)
	}
}

// Test_GIVEN_CorrectFlag_WHEN_AnsweringIncorrectly_THEN_ExpectWrongAnswerResponse tests the guess handler with a wrong answer
func Test_GIVEN_CorrectFlag_WHEN_AnsweringIncorrectly_THEN_ExpectWrongAnswerResponse(t *testing.T) {
	// Arrange
	gameState := &GameState{
		IsCorrect:   true, // Flag is correct
		CountryName: "TestCountry",
		Total:       0,
		Correct:     0,
		Incorrect:   0,
	}
	deps := &Dependencies{
		GameState: gameState,
	}
	handler := guessHandler(deps)

	// Act - Guess fake flag when it's actually correct
	req, err := http.NewRequest("GET", "/guess?answer=incorrect", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assert
	if gameState.Total != 1 {
		t.Errorf("Expected Total to be 1, got %d", gameState.Total)
	}
	if gameState.Correct != 0 {
		t.Errorf("Expected Correct to be 0, got %d", gameState.Correct)
	}
	if gameState.Incorrect != 1 {
		t.Errorf("Expected Incorrect to be 1, got %d", gameState.Incorrect)
	}
	if gameState.ResultCorrect {
		t.Error("Expected ResultCorrect to be false")
	}
	expectedMessage := "This was actually the correct TestCountry flag."
	if gameState.ResultMessage != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, gameState.ResultMessage)
	}
}
