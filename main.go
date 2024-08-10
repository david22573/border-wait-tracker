package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
)

func main() {
	go func() {
		feedWriter()
	}()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func feedWriter() {
	// Parse request
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL("https://www.bleepingcomputer.com/feed/")
	println(feed.Title)
}
