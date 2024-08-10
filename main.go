package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
)

type Entry struct {
	Name string
	Body string
}

func main() {
	URLS := []string{
		"https://bwt.cbp.gov/api/bwtRss/rssbyportnum/HTML/POV/250601",
		"https://bwt.cbp.gov/api/bwtRss/rssbyportnum/HTML/POV/250401",
	}
	go func() {
		for _, url := range URLS {
			feedWriter(url)
		}
	}()
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "CBP Wait Times",
		})
	})
	router.GET("/info", info)
	router.Run() // listen and serve on 0.0.0.0:8080
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

func info(c *gin.Context) {
	entries, err := getEntries()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	c.HTML(http.StatusOK, "info.tmpl", gin.H{
		"entries": entries,
	})
}

func getEntries() ([]Entry, error) {
	entries := []Entry{}
	files, err := os.ReadDir(".")
	if err != nil {
		return entries, fmt.Errorf("error reading directory: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			entry, err := loadFile(file.Name())
			if err != nil {
				return entries, fmt.Errorf("error reading file: %v", err)
			}
			entries = append(entries, Entry{
				Name: file.Name(),
				Body: string(entry),
			})
		}
	}
	return entries, nil
}

func loadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
