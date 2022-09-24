package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	strava "github.com/sgryczan/go.strava"
)

var (
	client = &http.Client{}
)

func GetURL(s string) (*string, error) {
	url, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: "GET",
		URL:    url,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return nil, err
	}
	log.Println(string(b))

	return nil, nil
}

func UploadRide(name string, data []byte) error {
	if tokenNeedsRefresh(currentAuth) {
		err := refreshAuthToken()
		if err != nil {
			return err
		}
	}
	client := strava.NewClient(currentAuth.AccessToken)
	service := strava.NewUploadsService(client)

	fmt.Printf("Uploading data...\n")

	upload, err := service.
		Create(strava.FileDataTypes.GPX, name, bytes.NewReader(data)).
		Private().
		Do()
	if err != nil {
		if e, ok := err.(strava.Error); ok && e.Message == "Authorization Error" {
			log.Printf("Make sure your token has 'write' permissions.")
			return err
		}

		// Ignore duplicate rides
		if e, ok := err.(strava.Error); ok && strings.Contains(e.Message, "duplicate") {
			log.Print(err)
			return nil
		}

		return err
	}

	log.Printf("Upload Complete...")

	log.Printf("Waiting a second so the upload will finish")
	time.Sleep(5 * time.Second)

	var jsonForDisplay []byte
	var processed bool
	var uploadSummary *strava.UploadDetailed
	for !processed {
		uploadSummary, err = service.Get(upload.Id).Do()
		if err != nil {
			return err
		}

		jsonForDisplay, _ = json.Marshal(uploadSummary)
		log.Printf(string(jsonForDisplay))

		if uploadSummary.ActivityId != 0 {
			processed = true
		} else {
			time.Sleep(time.Second * 5)
		}
	}

	log.Printf("Your new activity is id %d", uploadSummary.ActivityId)
	log.Printf("You can view it at http://www.strava.com/activities/%d", uploadSummary.ActivityId)

	return nil
}
