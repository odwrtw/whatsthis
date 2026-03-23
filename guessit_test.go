package guessit

import (
	"encoding/json"
	"os"
	"testing"
)

type Entry struct {
	FileName string `json:"filename"`
	Expected Guess  `json:"expected"`
}

func TestGuessMovies(t *testing.T) {
	filename := "tests/movies-testdata.json"
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Skipf("%s not available", filename)
	}
	var entries []Entry
	if err := json.Unmarshal(content, &entries); err != nil {
		t.Fatalf("invalid %s: %+v", filename, err)
	}
	for _, e := range entries {
		ourGuess := GuessIt(e.FileName)
		if ourGuess != e.Expected {
			t.Errorf("filename:%s\nOurs:%+v\nexpected:%+v\n", e.FileName, ourGuess, e.Expected)
			t.Fail()
		}
	}
}

func TestGuessTvShows(t *testing.T) {
	filename := "tests/tvshows-testdata.json"
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Skipf("%s not available", filename)
	}
	var entries []Entry
	if err := json.Unmarshal(content, &entries); err != nil {
		t.Fatalf("invalid %s: %+v", filename, err)
	}
	for _, e := range entries {
		ourGuess := GuessIt(e.FileName)
		if ourGuess != e.Expected {
			t.Errorf("filename:%s\nOurs:%+v\nexpected:%+v\n", e.FileName, ourGuess, e.Expected)
			t.Fail()
		}
	}
}
