package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/Pilfer/ultimate-guitar-scraper/pkg/ultimateguitar"
	"github.com/urfave/cli"
)

var FindTab = cli.Command{
	Name:        "find",
	Usage:       "ug find -title \"Hallelujah\" [-type chords|tabs|bass|ukulele|all]",
	Description: "Search by song name and print the lyrics/chords for the top-rated match",
	Aliases:     []string{"n"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "title",
			Usage: "Song title (and/or artist name) to search for",
		},
		cli.StringFlag{
			Name:  "type",
			Value: "chords",
			Usage: "Result type: chords, tabs, bass, ukulele, pro, official, or all",
		},
	},
	Action: findTab,
}

func findTab(c *cli.Context) {
	title := c.String("title")
	if title == "" {
		log.Fatal("You must provide a -title to search for, e.g. ug find -title \"Hallelujah\"")
	}

	s := ultimateguitar.New()
	result, err := s.Search(ultimateguitar.SearchParams{
		Title: title,
		Type:  []ultimateguitar.TabType{tabTypeFromString(c.String("type"))},
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(result.Tabs) == 0 {
		fmt.Println("No results found for:", title)
		return
	}

	// Pick the highest-voted result as a reasonable proxy for "best" match.
	best := result.Tabs[0]
	for _, tab := range result.Tabs {
		if tab.Votes > best.Votes {
			best = tab
		}
	}

	tab, err := s.GetTabByID(best.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("----------------------------------------------------------------------")
	fmt.Println("Song name:", tab.SongName, " by ", tab.ArtistName)
	fmt.Printf("(tab id %d, rating %.2f, %d votes — use `ug fetch -id %d` to grab this exact version again)\n", best.ID, best.Rating, best.Votes, best.ID)
	fmt.Println("----------------------------------------------------------------------")

	tabOut := strings.ReplaceAll(tab.Content, "[tab]", "")
	tabOut = strings.ReplaceAll(tabOut, "[/tab]", "")
	tabOut = strings.ReplaceAll(tabOut, "[ch]", "")
	tabOut = strings.ReplaceAll(tabOut, "[/ch]", "")
	fmt.Println(tabOut)
}

