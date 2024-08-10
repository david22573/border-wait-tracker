package main

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
)

type Entry struct {
	name string
	body []byte
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
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func feedWriter(url string) error {
	// Parse request
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)
	info := strings.Split(feed.Items[0].Description, "<br/>")[3]
	entry := &Entry{
		name: feed.Items[0].Title,
		body: []byte(info),
	}
	filename := strings.Split(feed.Items[0].Title, " ")[0]
	return os.WriteFile(filename, entry.body, 0600)
}
