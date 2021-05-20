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

func fetchDoc(u string) (string, string) {
	url := generateDocURL(u)

	log.Println("Fetching Google Doc: ", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", ErrorFetchingPage
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", ErrorReadingResponse
		}
		return string(contents), ""
	}
}

func preprocessDoc(doc string, id int) string {
	arr := [3]string{
		"\"",
		".",
		" ",
	}

	for i := 0; i < 10; i++ { // Iterate between 0 and 9
		cur := strconv.Itoa(i)
		for _, e := range arr {
			o := e + "c" + cur
			n := e + "c" + strconv.Itoa(id) + cur
			log.Println("Replacing ", o, " with ", n)
			doc = strings.ReplaceAll(doc, o, n)
		}
	}

	return doc
}

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
			if (row == 1) || (contents == "") { // Empty parent section, or header
				curPS.Contents = ""
			} else {
				doc, res := fetchDoc(contents)
				if res != "" {
					curPS.Contents = res
				} else {
					// Fix CSS selector problem
					curPS.Contents = preprocessDoc(doc, row)
				}
			}
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

func generateDocURL(u string) string {
	return u + "?embedded=true"
}

func getFrontPage(gsc string) ([]PageSection, string) {
	url := generateSheetURL(gsc)
	resp, err := http.Get(url)
	if err != nil {
		return nil, ErrorFetchingPage
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, ErrorReadingResponse
		}

		var newSheet GoogleSheet

		err = json.Unmarshal(contents, &newSheet)
		if err != nil {
			return nil, ErrorParsingPage
		}
		entries, ok := generateEntries(newSheet)
		if ok != "" {
			return nil, ok
		}

		return entries, ""
	}
}
