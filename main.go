package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
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
