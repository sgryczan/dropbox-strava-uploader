package dropbox

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/sgryczan/strava-uploader/pkg/models"
)

var (
	dropBoxAccessToken string
	DefaultPath        = "/Apps/tapiriik"
	client             = &http.Client{}
)

func ListFolderContents(s string) (*models.DropBoxListFolderResponse, error) {
	if s == "" {
		s = DefaultPath
	}

	url, err := url.Parse("https://api.dropboxapi.com/2/files/list_folder")
	if err != nil {
		log.Printf("[dropbox] [ListFolderContents] - error parsing url: %s\n", err)
		return nil, err
	}

	data := map[string]string{
		"path": "/Apps/tapiriik",
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("[dropbox] [ListFolderContents] - error marshalling data: %s\n", err)
		return nil, err
	}

	req := &http.Request{
		Method: "POST",
		URL:    url,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(string(body))),
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[dropbox] [ListFolderContents] - error performing request: %s\n", err)
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[dropbox] [ListFolderContents] - error reading response body: %s\n", err)
		return nil, err
	}

	/* rb, err := httputil.DumpRequest(req, true)
	log.Printf("[dropbox] [ListFolderContents] - request %s\n", rb)

	rb, err = httputil.DumpResponse(resp, true)
	log.Printf("[dropbox] [ListFolderContents] - response %s\n", rb) */

	listResponse := &models.DropBoxListFolderResponse{}
	err = json.Unmarshal(b, listResponse)
	if err != nil {
		log.Printf("[dropbox] [ListFolderContents] - error unmarshalling response body: %s\n", err)
		log.Printf("[dropbox] [ListFolderContents] - response body: %s\n", b)
		return nil, err
	}
	return listResponse, nil
}

func DownloadFile(p string) ([]byte, error) {
	url, err := url.Parse("https://content.dropboxapi.com/2/files/download")
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"path": p,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: "POST",
		URL:    url,
		Header: map[string][]string{
			"DropBox-API-Arg": {string(body)},
		},
	}

	resp, err := client.Do(req)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func MoveFile(fromPath, toPath string) error {
	log.Printf("moving file from %s to %s\n", fromPath, toPath)
	url, err := url.Parse("https://api.dropboxapi.com/2/files/move_v2")
	if err != nil {
		return err
	}

	data := map[string]string{
		"from_path": fromPath,
		"to_path":   toPath,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req := &http.Request{
		Method: "POST",
		URL:    url,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(string(body))),
	}

	resp, err := client.Do(req)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Print(string(b))

	return nil
}

func AuthIsGood() bool {
	t, _ := readTokenFile()
	if t == nil {
		return false
	}

	return true
}
