// Calling GitHub API
package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	// The ACCESS TOKEN. fmt.Println(tok.AccessToken)

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
	// initialize a client with OAuth
	domain := "poor.farm"
	http_client := oauthCall()
	stats, err := getStats(domain, http_client)

	if err != nil {
		log.Fatalf("error: %s", err)
	}

	fmt.Printf("%+v : %+v\n", domain, stats)
}
