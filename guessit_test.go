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

func TestGuess(t *testing.T) {
	content, err := os.ReadFile("testdata.json")
	if err != nil {
		t.Skip("testdata.json not available")
	}
	var entries []Entry
	if err := json.Unmarshal(content, &entries); err != nil {
		t.Fatalf("invalid testdata.json: %+v", err)
	}
	for _, e := range entries {
		ourGuess := GuessIt(e.FileName)
		if ourGuess != e.Expected {
			t.Errorf("filename:%s\nOurs:%+v\nexpected:%+v\n", e.FileName, ourGuess, e.Expected)
			t.Fail()
		}
	}
}
