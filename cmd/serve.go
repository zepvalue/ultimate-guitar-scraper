package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Pilfer/ultimate-guitar-scraper/pkg/ultimateguitar"
	"github.com/urfave/cli"
)

var ServeHTTP = cli.Command{
	Name:        "serve",
	Usage:       "ug serve -port 8080",
	Description: "Run a tiny HTTP server exposing /search and /find so you can query it remotely (e.g. via a tunnel or hosting platform)",
	Aliases:     []string{"http"},
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "port",
			Value: 8080,
			Usage: "Port to listen on. If a PORT env var is set (e.g. on Render), that takes priority.",
		},
		cli.StringFlag{
			Name:  "token",
			Usage: "Shared-secret token required on every request (as ?token=... or X-Auth-Token header). If empty, checks AUTH_TOKEN env var, else one is generated and printed on startup.",
		},
	},
	Action: serveHTTP,
}

func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func checkToken(r *http.Request, expected string) bool {
	if expected == "" {
		return true
	}
	if r.URL.Query().Get("token") == expected {
		return true
	}
	if r.Header.Get("X-Auth-Token") == expected {
		return true
	}
	return false
}

func serveHTTP(c *cli.Context) {
	port := c.Int("port")
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	token := c.String("token")
	if token == "" {
		token = os.Getenv("AUTH_TOKEN")
	}
	if token == "" {
		token = randomToken()
		fmt.Println("NOTE: no token set via -token or AUTH_TOKEN env var — generated one below.")
		fmt.Println("On a redeploy this token will change. Set AUTH_TOKEN as an env var on Render for a stable token.")
	}

	s := ultimateguitar.New()

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if !checkToken(r, token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		title := r.URL.Query().Get("title")
		if title == "" {
			http.Error(w, "missing ?title= param", http.StatusBadRequest)
			return
		}
		typeParam := r.URL.Query().Get("type")
		if typeParam == "" {
			typeParam = "chords"
		}

		result, err := s.Search(ultimateguitar.SearchParams{
			Title: title,
			Type:  []ultimateguitar.TabType{tabTypeFromString(typeParam)},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/find", func(w http.ResponseWriter, r *http.Request) {
		if !checkToken(r, token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		title := r.URL.Query().Get("title")
		if title == "" {
			http.Error(w, "missing ?title= param", http.StatusBadRequest)
			return
		}
		typeParam := r.URL.Query().Get("type")
		if typeParam == "" {
			typeParam = "chords"
		}

		result, err := s.Search(ultimateguitar.SearchParams{
			Title: title,
			Type:  []ultimateguitar.TabType{tabTypeFromString(typeParam)},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if len(result.Tabs) == 0 {
			http.Error(w, "no results found", http.StatusNotFound)
			return
		}

		best := result.Tabs[0]
		for _, tab := range result.Tabs {
			if tab.Votes > best.Votes {
				best = tab
			}
		}

		tab, err := s.GetTabByID(best.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		tabOut := strings.ReplaceAll(tab.Content, "[tab]", "")
		tabOut = strings.ReplaceAll(tabOut, "[/tab]", "")
		tabOut = strings.ReplaceAll(tabOut, "[ch]", "")
		tabOut = strings.ReplaceAll(tabOut, "[/ch]", "")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "%s by %s\n(tab id %d, rating %.2f, %d votes)\n\n%s\n", tab.SongName, tab.ArtistName, best.ID, best.Rating, best.Votes, tabOut)
	})

	fmt.Println("----------------------------------------------------------------------")
	fmt.Printf("Listening on http://0.0.0.0:%d\n", port)
	fmt.Printf("Auth token: %s\n", token)
	fmt.Println("Endpoints:")
	fmt.Printf("  GET /search?title=Hallelujah&token=%s\n", token)
	fmt.Printf("  GET /find?title=Hallelujah&token=%s\n", token)
	fmt.Println("----------------------------------------------------------------------")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

