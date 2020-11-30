package main

import (
	"errors"
	"fmt"
	"hackday-diversity/internal"
	"net/http"
	"net/url"
	"os"
	"time"
)

func GetJournalistsAndArticles(client *http.Client, db *internal.HackdayDB, apiKey string, date time.Time) error {
	dateStr := date.Format("2006-01-02")
	seed, _ := url.Parse(
		fmt.Sprintf(
			"https://content.guardianapis.com/search?use-date=first-publication&from-date=%s&to-date=%s&page-size=50&show-fields=bylineHtml&api-key=%s",
			dateStr, dateStr, apiKey,
		),
	)

	paginator := internal.NewCAPISearchPaginator(client, seed)

	for paginator.HasNext() {

		fmt.Printf("executing request: %s\n", paginator.URL)

		res, err := paginator.Next()
		if err != nil {
			return fmt.Errorf("pagination error: %w", err)
		}

		for _, result := range res.Response.Results {
			webUrl, err := url.Parse(result.WebURL)
			if err != nil {
				fmt.Printf("unable to parse raw URL %s\n", result.WebURL)
				continue
			}

			for _, info := range internal.JournalistInfoFromByline(result.Fields.BylineHTML, webUrl) {
				if err := db.InsertJournalist(info); err != nil {
					fmt.Printf("unable to insert journalist %s int db: %s\n", info.Name, err)
				}

				if err := db.InsertArticle(result.WebPublicationDate, result.ID, webUrl, info.Name); err != nil {
					fmt.Printf("unable to insert article %s into db: %s\n", webUrl, err)
				}
			}
		}
	}

	return nil
}

func main() {

	apiKey := os.Getenv("CAPI_API_KEY")
	if apiKey == "" {
		internal.ExitOnError(errors.New("CAPI_API_KEY not set"))
	}

	db, err := internal.NewDefaultHackdayDB()
	internal.ExitOnError(err)
	defer db.Close()

	client := &http.Client{}

	start := time.Date(2013, 10, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2018, 12, 31, 0, 0, 0, 0, time.Local)

	maxInflightDays := 5
	inflightDays := make(chan struct{}, maxInflightDays)

	for date := start; !date.After(end); date = date.Add(24 * time.Hour) {
		fmt.Printf("awaiting to process date %s\n", date)
		inflightDays <- struct{}{}
		go func(date time.Time) {
			defer func() { <-inflightDays }()
			fmt.Printf("running process for date %s\n", date)
			if err := GetJournalistsAndArticles(client, db, apiKey, date); err != nil {
				fmt.Printf("unable to process results: %s\n", err)
			}
		}(date)
	}

	for i := 0; i < maxInflightDays; i++ {
		fmt.Println("awaiting process to finish")
		inflightDays <- struct{}{}
	}
}
