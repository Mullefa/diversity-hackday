package main

import (
	"fmt"
	"hackday-diversity/internal"
	"io/ioutil"
	"net/http"
)

func GetJournalistImage(client *http.Client, db *internal.HackdayDB, info internal.JournalistInfo) error {
	if info.ProfileURL == nil || len(info.ImageName) > 0 {
		return nil
	}

	imageUrl, err := internal.ImageUrlFromProfileUrl(client, info.ProfileURL)
	if err != nil {
		if err == internal.NoProfileImage {
			return err
		}
		return fmt.Errorf("unable to get profile image url for journalist %s: %w", info.Name, err)
	}

	fmt.Printf("found %s: %s\n", info.Name, imageUrl)

	imageType, data, err := internal.DownloadImage(client, imageUrl)
	if err != nil {
		return fmt.Errorf("unable to download profile image for journalist: %s: %w", info.Name, err)
	}

	filename := fmt.Sprintf("./journalist-images/%s.%s", info.Name, imageType.Extension())
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("unable to save profile image for journalist %s to file: %w", info.Name, err)
	}
	if err := db.UpdateJournalistImage(info.Name, filename); err != nil {
		return fmt.Errorf("unable to update profile image for journalist %s in db: %w", info.Name, err)
	}
	return nil
}

func main() {

	db, err := internal.NewDefaultHackdayDB()
	internal.ExitOnError(err)
	defer db.Close()

	rows, err := db.GetJournalists()
	internal.ExitOnError(err)
	defer rows.Close()

	client := &http.Client{}

	maxInflightRequests := 3
	inflightRequests := make(chan struct{}, maxInflightRequests)

	for rows.Next() {
		info, err := rows.Scan()
		internal.ExitOnError(err)
		inflightRequests <- struct{}{}
		go func(info internal.JournalistInfo) {
			defer func() { <-inflightRequests }()
			if err := GetJournalistImage(client, db, info); err != nil && err != internal.NoProfileImage {
				fmt.Printf("unable to get journalist image: %s\n", err)
			}
		}(info)
	}

	internal.ExitOnError(rows.Err())

	for i := 0; i < maxInflightRequests; i++ {
		inflightRequests <- struct{}{}
	}
}
