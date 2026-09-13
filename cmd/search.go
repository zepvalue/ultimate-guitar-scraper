package cmd

import (
	"fmt"
	"log"

	"github.com/Pilfer/ultimate-guitar-scraper/pkg/ultimateguitar"
	"github.com/urfave/cli"
)

var SearchTabs = cli.Command{
	Name:        "search",
	Usage:       "ug search -title {songName} [-type chords|tabs|bass|ukulele|all]",
	Description: "Search ultimate-guitar.com for tabs/chords by song name",
	Aliases:     []string{"s"},
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
	Action: searchTabs,
}

func tabTypeFromString(s string) ultimateguitar.TabType {
	switch s {
	case "tabs":
		return ultimateguitar.TabTypeTabs
	case "bass":
		return ultimateguitar.TabTypeBass
	case "ukulele":
		return ultimateguitar.TabTypeUkulele
	case "pro":
		return ultimateguitar.TabTypePro
	case "official":
		return ultimateguitar.TabTypeOfficial
	case "all":
		return ultimateguitar.TabTypeAll
	default:
		return ultimateguitar.TabTypeChords
	}
}

func searchTabs(c *cli.Context) {
	title := c.String("title")
	if title == "" {
		log.Fatal("You must provide a -title to search for, e.g. ug search -title \"Hallelujah\"")
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

	fmt.Println("----------------------------------------------------------------------")
	fmt.Printf("%-10s %-30s %-25s %-10s %s\n", "ID", "Song", "Artist", "Rating", "Votes")
	fmt.Println("----------------------------------------------------------------------")
	for _, tab := range result.Tabs {
		fmt.Printf("%-10d %-30s %-25s %-10.2f %d\n", tab.ID, truncate(tab.SongName, 30), truncate(string(tab.ArtistName), 25), tab.Rating, tab.Votes)
	}
	fmt.Println("----------------------------------------------------------------------")
	fmt.Println("Use `ug fetch -id <ID>` to pull the lyrics/chords for a specific result.")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

