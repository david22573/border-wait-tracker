package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

var (
	URLS = []string{
		"https://bwt.cbp.gov/api/bwtRss/rssbyportnum/HTML/POV/250601",
		"https://bwt.cbp.gov/api/bwtRss/rssbyportnum/HTML/POV/250401",
	}
)

func FetchTimes() {
	go func() {
		for {
			for _, url := range URLS {
				feedWriter(url)
				time.Sleep(time.Minute * 1)
			}
			time.Sleep(time.Minute * 5)
		}
	}()
}

func feedWriter(url string) error {
	// Parse request
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)
	info := strings.Split(feed.Items[0].Description, "<br/>")[3]
	entry := &Entry{
		Name: feed.Items[0].Title,
		Body: info,
	}
	cleanName, _ := cleanFilename(entry.Name)
	filename := cleanName + ".txt"
	return os.WriteFile(filename, []byte(entry.Body), 0600)
}

func cleanFilename(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}
	lower := strings.ToLower(name)
	r := strings.ReplaceAll(lower, " - ", " ")
	cleanName := strings.ReplaceAll(r, " ", "-")
	return cleanName, nil
}
