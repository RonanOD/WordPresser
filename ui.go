package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
	"sort"
	"sync"
)

// SafeSiteData Map of Urls Strings to SiteDataModel with mutex for thread safety
type SafeSiteData struct {
	mu sync.Mutex
	siteData  map[string]SiteDataModel
}

type SiteDataModel struct {
	description string
	views       []float64
}

// Adding stat strings to the map
func (c *SafeSiteData) setSiteData(url string, data SiteDataModel) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map
	c.siteData[url] = data
	c.mu.Unlock()
}

// Value get a value from map
func (c *SafeSiteData) Value(url string) SiteDataModel {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map
	defer c.mu.Unlock()
	return c.siteData[url]
}

// keys get a sorted list of urls
func (c *SafeSiteData) urls() []string {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map
	defer c.mu.Unlock()

	keys := make([]string, len(c.siteData))

	i := 0
	for k := range c.siteData {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// get formatted string of stat struct
func (stat Stat) String() string {
	siteDetails := fmt.Sprintf("Stats:\n")
	siteDetails += fmt.Sprintf("\nToday:\t Views: %d\t Visitors: %d\n",
		stat.ViewsToday, stat.VisitorsToday)
	siteDetails += fmt.Sprintf("Yesterday:\t Views: %d\t Visitors: %d\n",
		stat.ViewsYesterday, stat.VisitorsYesterday)
	return siteDetails
}

// IsAllZeros termUI has an issue if data in bar chart data is all zeros it
// has a memory leak, this method checks if we've all zeros before rendering
// https://github.com/gizak/termui/issues/245
func IsAllZeros(data []float64) bool {
	if len(data) == 0 {
		return false
	}
	for _, view := range data {
		if view != 0 {
			return false
		}
	}
	return true
}

// ExtractFloatArrayFromStats the json visits data is unmarshalled into an empty interface, this method
// converts it into a float slice
func ExtractFloatArrayFromStats(stats [][3]interface{}) []float64 {
	views := make([]float64, 30)

	for i := 0; i < 30; i++ {
		view, ok := stats[i][1].(float64)
		if ok {
			views[i] = view
		} else {
			log.Fatalf("Failed to convert views data to float")
		}
	}
	return views
}

// InitUIElements populate initial state of ui
func InitUIElements(safeSiteData *SafeSiteData, siteList *widgets.List, selectedBox *widgets.Paragraph, barChart *widgets.BarChart) {
	titleBox := widgets.NewParagraph()
	titleBox.Title = "WordPresser"
	titleBox.Text = "Press up/down keys to scroll list, esc to exit."
	titleBox.SetRect(0, 0, 85, 3)
	titleBox.BorderStyle.Fg = ui.ColorCyan

	ui.Render(titleBox)


	siteList.Title = "List"
	siteList.Rows = safeSiteData.urls()
	siteList.TextStyle = ui.NewStyle(ui.ColorYellow)
	siteList.WrapText = false
	siteList.SetRect(0, 3, 45, 13)

	ui.Render(siteList)

	selectedBox.WrapText = true
	selectedBox.SetRect(45, 3, 85, 13)

	ui.Render(selectedBox)

	barChart.Title = "Views last 20 days"
	barChart.SetRect(0, 13, 85, 23)
	barChart.BarWidth = 3
	barChart.BarColors = []ui.Color{ui.ColorCyan}
	barChart.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorBlue)}
	barChart.NumStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}

	ui.Render(barChart)
}

// ListenForKeyboardEvents handle keyboard events
func ListenForKeyboardEvents(safeSiteData *SafeSiteData, siteList *widgets.List,
								selectedBox *widgets.Paragraph, barChart *widgets.BarChart) {
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "<Escape>":
			return
		case "<Down>":
			siteList.ScrollDown()
		case "<Up>":
			siteList.ScrollUp()
		}
		data := safeSiteData.Value(siteList.Rows[siteList.SelectedRow])
		selectedBox.Text = data.description

		views := data.views[len(data.views)-20:]
		if !IsAllZeros(views) {
			barChart.Data = views
		}

		ui.Render(selectedBox)
		ui.Render(siteList)
		ui.Render(barChart)

	}
}



