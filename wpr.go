// Utility to query the stats of all of your WordPress sites from the command line.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
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
func oauthCall() (*http.Client, string) {
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
	fmt.Println("=================================================================================")
	fmt.Println("No OAuth access token. You have to visit your wordpress site to grant permission.")
	fmt.Println("Click on the URL below, grant permission and copy the \"code\" parameter from the")
	fmt.Println("site address you are redirected to. Paste it into this console and hit enter.")
	fmt.Println("=================================================================================")
	fmt.Printf("URL: %v\n> ", init_url)

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

	return client, tok.AccessToken
}

// Return a Sites object for a given site ID
func getStats(id string, http_client *http.Client, token string) (*Sites, error) {
	// HTTP call
		url := fmt.Sprintf("https://public-api.wordpress.com/rest/v1.1/sites/%s/stats", id)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", token))
	res, err := http_client.Do(req)
	if err != nil {
		log.Fatalf("error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		/* How to get bytes as string.
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		*/
		// Decode JSON
		sites := &Sites{}
		dec := json.NewDecoder(res.Body)
		if err := dec.Decode(sites); err != nil {
			return nil, err
		}

		return sites, nil
	}
	return nil, nil
}

// Get a list of sites to return stats for.
func getSites(http_client *http.Client, request *http.Request, token string) (*AllSites, error) {
	request, _ = http.NewRequest("GET", "https://public-api.wordpress.com/rest/v1/me/sites", nil)
	request.Header.Set("authorization", fmt.Sprintf("Bearer %s", token))
	res, err := http_client.Do(request)
	if err != nil {
		log.Fatalf("error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		sites := &AllSites{}
		dec := json.NewDecoder(res.Body)
		if err := dec.Decode(sites); err != nil {
			log.Fatalf("error: %s", err)
			return nil, err
		}
		return sites, nil
	} else {
		err = fmt.Errorf("Could not get list. HTTP Status code: %d", res.StatusCode)
		return nil, err
	}
}


func lookupSiteStats(http_client *http.Client, token string, site Site, safeSiteData SafeSiteData, siteList *widgets.List, selectedBox *widgets.Paragraph) {
	// fetch site stats from rest api
	stats, err := getStats(fmt.Sprint(site.ID), http_client, token)

	if err != nil {
		log.Fatalf("error: %s", err)
	} else {
		// populate site data map with results
		safeSiteData.setSiteData(site.URL, stats.Stats.String())
		//if stats are for currently selected row in list refresh ui
		if siteList.Rows[siteList.SelectedRow] == site.URL {
			selectedBox.Text = safeSiteData.Value(site.URL)
			ui.Render(selectedBox)
		}
	}
}

// Main function that executes everything.
func main() {
	token_file := ".token"
	var http_client *http.Client
	var request *http.Request
	var token string

	// Set up http client. Might need to authenticate.
	if _, err := os.Stat(token_file); os.IsNotExist(err) {
		// initialize a client with OAuth. Drops .token for next time.
		http_client, token = oauthCall()
	} else {
		// load a client with bearer token
		content, err := ioutil.ReadFile(token_file)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		// Convert []byte to string
		token = string(content)
		http_client = &http.Client{}
	}
	// Display results. Use a channel to store each separate go call
	allSites, err := getSites(http_client, request, token)
	if err != nil {
		log.Fatalf("error: %s", err)
	} else {
		if err := ui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
		}
		defer ui.Close()

		safeSiteData := SafeSiteData{siteData: make(map[string]string)}
		siteList := widgets.NewList()
		selectedBox := widgets.NewParagraph()

		//Prepopulate site data with fetching strings
		for _, currentSite := range allSites.Sites {
			safeSiteData.setSiteData(currentSite.URL, "Fetching Data for "+currentSite.URL)
		}

		InitUIElements(safeSiteData, siteList, selectedBox)

		for _, site := range allSites.Sites {
			go lookupSiteStats(http_client, token, site, safeSiteData, siteList, selectedBox)
		}

		ListenForKeyboardEvents(safeSiteData, siteList, selectedBox)

	}
}


