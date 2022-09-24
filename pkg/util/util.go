package util

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/sgryczan/strava-uploader/pkg/dropbox"
	"github.com/sgryczan/strava-uploader/pkg/strava"
)

func StartPeriodicCollection(id string, t time.Duration) {
	for {
		CollectAndUpload(id)
		time.Sleep(t)
	}
}

func CollectAndUpload(id string) {
	listResponse, err := dropbox.ListFolderContents(dropbox.DefaultPath)
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	log.Printf("Found %d entries.", len(listResponse.Entries))
	for i, entry := range listResponse.Entries {
		log.Printf("processing entry %d of %d: %s", i+1, len(listResponse.Entries), entry.PathDisplay)
		if entry.Type == "folder" {
			continue
		}

		// Download the file
		b, err := dropbox.DownloadFile(entry.PathDisplay)
		if err != nil {
			log.Fatalf("err: %s", err.Error())
		}

		// If we dont have a refresh token, lets do the oauth2 flow
		var notifiedUser bool
		for !strava.AuthTokenExists() {
			if !notifiedUser {
				log.Printf("no strava auth token found. Please navigate to oauth.czan.io:8321 and get one.\n")
				notifiedUser = true
			}

			log.Printf("sleeping\n")
			time.Sleep(time.Second * 10)
		}

		// Attempt to upload to strava
		err = strava.UploadRide(entry.Name, b)
		if err != nil {
			log.Print(err.Error())
			os.Exit(1)
		}

		// Move the file to the processed folder
		err = dropbox.MoveFile(entry.PathDisplay, strings.Replace(entry.PathDisplay, "/Apps/tapiriik", "/Apps/tapiriik/processed", -1))
		if err != nil {
			log.Print(err.Error())
			os.Exit(1)
		}
	}
	log.Printf("Successfully processed all files.\n")
}
