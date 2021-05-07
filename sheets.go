package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type GoogleSheetCell struct {
	Value string `json:"$t"` // Unsure whether to use $t or inputValue (they're the same AFAIK)
	Row   string `json:"row"`
	Col   string `json:"col"`
}

type GoogleSheetEntry struct {
	Cell GoogleSheetCell `json:"gs$cell"`
}

type GoogleSheetFeed struct {
	Entries []GoogleSheetEntry `json:"entry"`
	// Updated string             `json:"updated"` // needs to be changed to the object it is (has subclass $t)
}

type GoogleSheet struct {
	Feed GoogleSheetFeed `json:"feed"`
}

func generateSheetURL(sheetID string) string {
	return "https://spreadsheets.google.com/feeds/cells/" + sheetID + "/1/public/full?alt=json"
}

func ParseStructurePage(gsc string) (string, string) {
	url := generateSheetURL(gsc)
	resp, err := http.Get(url)
	if err != nil {
		return "", ErrorFetchingPage
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", ErrorReadingResponse
		}

		var newSheet GoogleSheet

		err = json.Unmarshal(contents, &newSheet)
		if err != nil {
			panic(err)
			// return "", ErrorParsingPage
		}
		res, err := json.Marshal(newSheet)
		if err != nil {
			panic(err)
		}
		return string(res), ""
	}
	// return url, ""
}
