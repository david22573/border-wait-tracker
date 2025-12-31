package main

import (
        "flag"
        "fmt"
        "net/http"
        "os"
        "strings"
        "sync"
        "time"

        "github.com/gin-gonic/gin"
        "github.com/mmcdole/gofeed"
)

type Entry struct {
        Name string
        Body string
}

var (
        urls = []string{
                // Otay Mesa (250601)
                "https://bwt.cbp.gov/api/bwtRss/rssbyportnum/HTML/POV/250601",
                "https://bwt.cbp.gov/api/bwtRss/rssbyportnum/HTML/PED/250601",

                // San Ysidro (250401) - Note: Feed titles say "San Ysidro", not "Otay General"
                "https://bwt.cbp.gov/api/bwtRss/rssbyportnum/HTML/POV/250401",
                "https://bwt.cbp.gov/api/bwtRss/rssbyportnum/HTML/PED/250401",
        }

        // In-memory storage for entries, protected by mutex
        mu      sync.RWMutex
        entries = make(map[string]string) // key: cleaned name, value: body
)

func main() {
        flag.Parse()
        args := flag.Args()

        if len(args) == 0 {
                // Default behavior: run fetch once and output results
                if err := fetchAll(); err != nil {
                        fmt.Printf("Error fetching data: %v\n", err)
                        os.Exit(1)
                }
                printEntries()
                return
        }

        switch args[0] {
        case "server":
                runServer()
        default:
                fmt.Println("Unknown command. Use:")
                fmt.Println("  myapp          # run fetch once and print")
                fmt.Println("  myapp server   # start HTTP server")
                os.Exit(1)
        }
}

func runServer() {
        // Initial fetch
        if err := fetchAll(); err != nil {
                fmt.Printf("Initial fetch error: %v\n", err)
        }

        // Start background fetcher
        go startBackgroundFetcher()

        router := gin.Default()
        router.LoadHTMLGlob("templates/*")

        router.GET("/", func(c *gin.Context) {
                c.HTML(http.StatusOK, "index.tmpl", gin.H{
                        "title": "CBP Wait Times",
                })
        })

        router.GET("/info", info)

        router.Run()
}

func startBackgroundFetcher() {
        ticker := time.NewTicker(5 * time.Minute)
        for range ticker.C {
                if err := fetchAll(); err != nil {
                        fmt.Printf("Background fetch error: %v\n", err)
                }
        }
}

func fetchAll() error {
        var wg sync.WaitGroup
        var fetchErr error
        mu.Lock() // Lock for writing new entries
        defer mu.Unlock()

        newEntries := make(map[string]string)
        errChan := make(chan error, len(urls))

        for _, url := range urls {
                wg.Add(1)
                go func(u string) {
                        defer wg.Done()
                        name, body, err := fetchSingle(u)
                        if err != nil {
                                errChan <- err
                                return
                        }
                        newEntries[name] = body
                }(url)
        }

        wg.Wait()
        close(errChan)

        for err := range errChan {
                if fetchErr == nil {
                        fetchErr = err
                } else {
                        fetchErr = fmt.Errorf("%v; %w", fetchErr, err)
                }
        }

        if fetchErr != nil {
                return fetchErr
        }

        entries = newEntries // Atomic update
        return nil
}

func fetchSingle(url string) (string, string, error) {
        fp := gofeed.NewParser()
        feed, err := fp.ParseURL(url)
        if err != nil {
                return "", "", fmt.Errorf("parse error for %s: %w", url, err)
        }

        if len(feed.Items) == 0 {
                return "", "", fmt.Errorf("no RSS items found for %s", url)
        }

        parts := strings.Split(feed.Items[0].Description, "<br/>")
        if len(parts) < 4 {
                return "", "", fmt.Errorf("unexpected item format for %s", url)
        }

        info := parts[3]
        cleanName, err := cleanFilename(feed.Items[0].Title)
        if err != nil {
                return "", "", err
        }

        // Extract lane type from URL (POV or PED) and append for uniqueness
        urlParts := strings.Split(url, "/")
        laneType := strings.ToLower(urlParts[len(urlParts)-2])

        return cleanName + "-" + laneType, info, nil
}

func cleanFilename(name string) (string, error) {
        if name == "" {
                return "", fmt.Errorf("filename cannot be empty")
        }
        name = strings.ToLower(name)
        name = strings.ReplaceAll(name, " - ", " ")
        name = strings.ReplaceAll(name, " ", "-")
        return name, nil
}

func info(c *gin.Context) {
        mu.RLock()
        defer mu.RUnlock()

        var entryList []Entry
        for name, body := range entries {
                entryList = append(entryList, Entry{
                        Name: name + ".txt", // Mimic original filename for compatibility
                        Body: body,
                })
        }

        c.HTML(http.StatusOK, "info.tmpl", gin.H{
                "entries": entryList,
        })
}

func printEntries() {
        mu.RLock()
        defer mu.RUnlock()

        for name, body := range entries {
                fmt.Printf("=== %s ===\n%s\n\n", name, body)
        }
}