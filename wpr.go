// Calling GitHub API
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// Sites represents all user sites.
// TODO: Add visits struct for graphing: https://github.com/gizak/termui
type Sites struct {
	Date  string `json:"date"`
	Stats Stat   `json:"stats"`
}

// Stat object for a given site.
type Stat struct {
	VisitorsToday     int `json:"visitors_today"`
	VisitorsYesterday int `json:"visitors_yesterday"`
	ViewsToday        int `json:"views_today"`
	ViewsYesterday    int `json:"views_yesterday"`
}

// Global list of sites
type AllSites struct {
	Sites []Site `json:"sites"`
}

type Site struct {
	ID  int    `json:"ID"`
	URL string `json:"URL"`
}

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

// Make OAuth call to authorize the app.
func oauthCall() *http.Client {
	ctx := context.Background()
	redirect_uri := goDotEnvVariable("REDIRECT_URI")
	conf := &oauth2.Config{
		ClientID:     goDotEnvVariable("CLIENT_ID"),
		ClientSecret: goDotEnvVariable("CLIENT_SECRET"),
		Scopes:       []string{"global"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://public-api.wordpress.com/oauth2/authorize",
			TokenURL: "https://public-api.wordpress.com/oauth2/token",
		},
		RedirectURL: redirect_uri,
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	init_url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", init_url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	// Write out the token so we can reuse token next time.
	io_err := ioutil.WriteFile(".token", []byte(tok.AccessToken), 0644)
	if io_err != nil {
		log.Fatal(io_err)
	}

	if err != nil {
		log.Fatal(err)
	}
	client := conf.Client(ctx, tok)

	return client
}

// Return a Sites object for a given domain
func getStats(domain string, http_client *http.Client) (*Sites, error) {
	// HTTP call
	url := fmt.Sprintf("https://public-api.wordpress.com/rest/v1.1/sites/%s/stats", domain)

	resp, err := http_client.Get(url)
	if err != nil {
		log.Fatalf("error: %s", err)
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		/* How to get bytes as string.
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		*/
		// Decode JSON
		sites := &Sites{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(sites); err != nil {
			return nil, err
		}

		return sites, nil
	}
	return nil, nil
}

func main() {
	domain := "poor.farm"
	token_file := ".token"
	var http_client *http.Client
	if _, err := os.Stat(token_file); os.IsNotExist(err) {
		// initialize a client with OAuth. Drops .token for next time.
		http_client = oauthCall()
	} else {
		// load a client with bearer token
		content, err := ioutil.ReadFile(token_file)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		// Convert []byte to string and print to screen
		token := string(content)
		http_client = &http.Client{}
		req, _ := http.NewRequest("GET", "https://public-api.wordpress.com/rest/v1/me/sites", nil)
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", token))
		res, err := http_client.Do(req)
		if err != nil {
			log.Fatalf("error: %s", err)
			return
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			sites := &AllSites{}
			dec := json.NewDecoder(res.Body)
			if err := dec.Decode(sites); err != nil {
				log.Fatalf("error: %s", err)
			}
			fmt.Printf("%+v\n", sites)
		}
		return
	}

	stats, err := getStats(domain, http_client)

	if err != nil {
		log.Fatalf("error: %s", err)
	}

	fmt.Printf("%+v : %+v\n", domain, stats)
}
