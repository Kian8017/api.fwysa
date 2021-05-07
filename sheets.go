package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const DEPTH_CHARACTER string = "~"

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

/*
type PageSection struct {
        Parent      int    `json:"parent"`
        ID          int    `json:"id"`
        Title       string `json:"title"`
        Expanded    string   `json:"expanded"`
        Highlighted string   `json:"highlighted"`
        Contents    string `json:"contents"`
}
*/

func generateEntries(gs GoogleSheet) ([]PageSection, string) {
	ps := make(map[int]PageSection) // row -> PageSection

	lastSeen := make(map[int]int) // depth -> index (row)

	for _, currentEntry := range gs.Feed.Entries {
		contents := currentEntry.Cell.Value

		row, err := strconv.Atoi(currentEntry.Cell.Row)
		if err != nil {
			return nil, ErrorParsingIndex
		}

		col, err := strconv.Atoi(currentEntry.Cell.Col)
		if err != nil {
			return nil, ErrorParsingIndex
		}

		curPS, ok := ps[row]
		if !ok {
			curPS = PageSection{}
			curPS.ID = row
		}

		switch col {
		case 1: // Section Name
			curPS.Title = strings.ReplaceAll(contents, DEPTH_CHARACTER, "")
			depth := strings.Count(contents, DEPTH_CHARACTER)

			// Set lastSeen
			lastSeen[depth] = row
			if depth > 0 { // Since 0 is the nil value for an int, shouldn't need to set it for those with depth 0
				curPS.Parent = lastSeen[depth-1]
			}
		case 2: // Expanded?
			curPS.Expanded = contents
		case 3: // Highlighted?
			curPS.Highlighted = contents
		case 4: // Google Doc Link
			// FIXME: Fetch contents here...
			curPS.Contents = contents
		default:
			log.Fatal("Unknown column, ", col)
		}

		ps[row] = curPS
	}

	var p []PageSection

	for _, cps := range ps {
		if cps.ID == 1 { // Discard header row
			continue
		}
		p = append(p, cps)
	}

	// iterate over ps, create p
	return p, ""
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
		entries, ok := generateEntries(newSheet)
		if ok != "" {
			return "", ok
		}

		res, err := json.Marshal(entries)
		if err != nil {
			panic(err)
		}
		return string(res), ""
	}
	// return url, ""
}
