package guessit

import (
	"encoding/json"
	"os"
	"testing"
)

type entry struct {
	FileName string `json:"filename"`
	Expected Guess  `json:"expected"`
}

func testGuessFromFile(t *testing.T, filename string) {
	t.Helper()

	content, err := os.ReadFile(filename) //nolint:gosec // test data paths are not user-controlled
	if err != nil {
		t.Skipf("%s not available", filename)
	}

	var entries []entry
	if err := json.Unmarshal(content, &entries); err != nil {
		t.Fatalf("invalid %s: %+v", filename, err)
	}

	for _, e := range entries {
		got := GuessIt(e.FileName)
		if got != e.Expected {
			t.Errorf("filename: %s\ngot:      %+v\nexpected: %+v", e.FileName, got, e.Expected)
		}
	}
}

func TestGuessMovies(t *testing.T) {
	testGuessFromFile(t, "testdata/movies-testdata.json")
}

func TestGuessTvShows(t *testing.T) {
	testGuessFromFile(t, "testdata/tvshows-testdata.json")
}
