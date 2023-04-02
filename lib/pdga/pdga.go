package pdga

import (
	"fmt"
	"log"

	"errors"
	"strings"

	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Player struct {
	ID             string
	name           string
	location       string
	classification string
	currentRating  string
	doc            string //HTML doc to cache
}

// Constructor for Player
func NewPlayer(pdgaID string) (*Player, error) {
	name, currentRating, location, classification, err := grabBasicPdgaInfo(pdgaID)
	if err == nil {
		return &Player{
			ID:             pdgaID,
			name:           name,
			location:       location,
			currentRating:  currentRating,
			classification: classification,
		}, nil
	}
	return nil, err
}

// Info returns basic information about a pdga player
func (p Player) Info() string {
	infoMessage := `Here is the basic information on PDGA Number %s.
	Name:           %s
	Classification: %s
	Location: 		%s
	Current Rating: %s
	`
	return fmt.Sprintf(infoMessage, p.ID, p.name, p.classification, p.location, p.currentRating)
}

// PredictRating calculates your expected Rating update
func (p Player) PredictRating() string {
	return fmt.Sprintf("Your predicted rating is 1050.  You're a crusher!")
}

func grabBasicPdgaInfo(id string) (name, currentRating, location, classification string, err error) {
	resp, err := http.Get("https://www.pdga.com/player/" + id)
	if err != nil {
		log.Println("Error retrieving pdga profile:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			err = errors.New("PDGA Number invalid")
		} else {
			errorString := fmt.Sprintf("status code error: %d %s", resp.StatusCode, resp.Status)
			err = errors.New(errorString)
		}
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	// Find the players name
	stringRemoval := "#" + id
	// Get players name from page title, strip out ID from string and trim
	name = strings.TrimSpace(strings.ReplaceAll(doc.Find("#page-title").Each(func(i int, s *goquery.Selection) {}).Text(), stringRemoval, ""))

	// Find our rating, location, classification
	findInList := func(doc *goquery.Document, selector string) string {
		return strings.TrimSpace(doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			s.Find("strong").Remove()
			s.Find("small").Remove()
		}).Text())
	}
	currentRating = findInList(doc, ".current-rating")
	classification = findInList(doc, ".classification")
	location = findInList(doc, ".location")

	return
}
