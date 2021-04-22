package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"sort"
	"sync"
)

// SafeSiteData Map of Urls Strings to Stats stings with mutex for thread safety
type SafeSiteData struct {
	mu sync.Mutex
	siteData  map[string]string
}

// Adding stat strings to the map
func (c *SafeSiteData) setSiteData(site, data string) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map
	c.siteData[site] = data
	c.mu.Unlock()
}

// Value get a value from map
func (c *SafeSiteData) Value(key string) string {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map
	defer c.mu.Unlock()
	return c.siteData[key]
}

// keys get a sorted list of urls
func (c *SafeSiteData) keys() []string {
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

// InitUIElements populate initial state of ui
func InitUIElements(safeSiteData SafeSiteData, siteList *widgets.List, selectedBox *widgets.Paragraph) {
	titleBox := widgets.NewParagraph()
	titleBox.Title = "WordPresser"
	titleBox.Text = "Press up/down keys to scroll list, esc to exit."
	titleBox.SetRect(0, 0, 75, 3)
	titleBox.BorderStyle.Fg = ui.ColorCyan

	ui.Render(titleBox)


	siteList.Title = "List"
	siteList.Rows = safeSiteData.keys()
	siteList.TextStyle = ui.NewStyle(ui.ColorYellow)
	siteList.WrapText = false
	siteList.SetRect(0, 3, 35, 13)

	ui.Render(siteList)

	selectedBox.WrapText = true
	selectedBox.SetRect(35, 3, 75, 13)

	ui.Render(selectedBox)


}

// ListenForKeyboardEvents handle keyboard events
func ListenForKeyboardEvents(safeSiteData SafeSiteData, siteList *widgets.List, selectedBox *widgets.Paragraph) {
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
		selectedBox.Text = safeSiteData.Value(siteList.Rows[siteList.SelectedRow])

		ui.Render(selectedBox)
		ui.Render(siteList)
	}
}

