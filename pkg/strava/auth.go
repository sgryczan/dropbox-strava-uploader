// oauth_example.go provides a simple example implementing Strava OAuth
// using the go.strava library.
//
// usage:
//   > go get github.com/strava/go.strava
//   > cd $GOPATH/github.com/strava/go.strava/examples
//   > go run oauth_example.go -id=youappsid -secret=yourappsecret
//
//   Visit http://localhost:8080 in your webbrowser
//
//   Application id and secret can be found at https://www.strava.com/settings/api
package strava

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	strava "github.com/sgryczan/go.strava"
)

const port = 8321 // port of local demo server

var (
	authenticator *strava.OAuthAuthenticator
	currentAuth   *strava.AuthorizationResponse
	Debug bool
)

func init() {
	cid, err := strconv.Atoi(os.Getenv("STRAVA_CLIENT_ID"))
	if err != nil {
		log.Fatalf("%s", err)
	}
	strava.ClientId = cid
	strava.ClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")

	readTokenFile()
}

type RefreshTokenResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int    `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
}

func AuthTokenExists() bool {
	if currentAuth == nil {
		return false
	}

	return true
}

func StartAuthServer() {
	if strava.ClientId == 0 || strava.ClientSecret == "" {
		fmt.Println("\nPlease provide your application's client_id and client_secret.")
		fmt.Println(" ")

		flag.PrintDefaults()
		os.Exit(1)
	}

	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://oauth.czan.io:%d/exchange_token", port),
		RequestClientGenerator: nil,
	}

	http.HandleFunc("/", indexHandler)

	path, err := authenticator.CallbackPath()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	http.HandleFunc(path, authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))

	// start the server
	fmt.Printf("Visit http://oauth.czan.io:%d/ to get a token\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// you should make this a template in your real application
	fmt.Fprintf(w, `<a href="%s">`, authenticator.AuthorizationURL("state1", strava.Permissions.ReadWriteActivity, true))
	fmt.Fprint(w, `<img src="http://strava.github.io/api/images/ConnectWithStrava.png" />`)
	fmt.Fprint(w, `</a>`)
}

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "State: %s\n\n", auth.State)
	fmt.Fprintf(w, "Access Token: %s\n\n", auth.AccessToken)

	fmt.Fprintf(w, "User Details:\n")
	content, _ := json.MarshalIndent(auth.Athlete, "", " ")
	fmt.Fprint(w, string(content))
	fmt.Fprintf(w, fmt.Sprintf("%+v", auth))

	currentAuth = auth
	writeTokenFile(auth)
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Authorization Failure:\n")

	// some standard error checking
	if err == strava.OAuthAuthorizationDeniedErr {
		fmt.Fprint(w, "The user clicked the 'Do not Authorize' button on the previous page.\n")
	} else if err == strava.OAuthInvalidCredentialsErr {
		fmt.Fprint(w, "You provided an incorrect client_id or client_secret.\n")
	} else if err == strava.OAuthInvalidCodeErr {
		fmt.Fprint(w, "The temporary token was not recognized")
	} else if err == strava.OAuthServerErr {
		fmt.Fprint(w, "There was some sort of server error")
	} else {
		fmt.Fprint(w, err)
	}
}

func readTokenFile() error {
	auth := &strava.AuthorizationResponse{}
	data, err := os.ReadFile("/app/token.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, auth)
	if err != nil {
		return err
	}

	currentAuth = auth
	return nil
}

func writeTokenFile(auth *strava.AuthorizationResponse) error {
	b, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	err = os.WriteFile("/app/token.json", b, 0644)
	if err != nil {
		return err
	}

	log.Println("wrote auth token to file")
	return nil
}

func tokenNeedsRefresh(token *strava.AuthorizationResponse) bool {
	// Tokens are valid for 6 hours, become refresh-able when they expire within 1 hour.
	now := time.Now()
	refreshTime := now.Add(time.Hour * 1).Unix()

	if Debug {
		log.Printf("auth token refreshTime: %d\n", refreshTime)
		log.Printf("auth token expiry time: %d\n", currentAuth.ExpiresAt)
	}

	if refreshTime > int64(currentAuth.ExpiresAt) {
		log.Printf("token is expired!\n")
		return true
	}

	if Debug {
		validFor := int64(currentAuth.ExpiresAt) - refreshTime
		log.Printf("token is valid for %d seconds\n", validFor)
	}
	
	return false
}

func refreshAuthToken() error {
	log.Printf("Refreshing oauth token...\n")
	url, err := url.Parse("https://www.strava.com/api/v3/oauth/token")
	if err != nil {
		return err
	}

	data := map[string]string{
		"client_id":     strconv.Itoa(strava.ClientId),
		"client_secret": strava.ClientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": currentAuth.RefreshToken,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req := &http.Request{
		Method: "POST",
		URL:    url,
		Body:   io.NopCloser(strings.NewReader(string(body))),
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	tokenResponse := &strava.AuthorizationResponse{}
	err = json.Unmarshal(b, tokenResponse)
	if err != nil {
		return err
	}

	currentAuth = tokenResponse

	b, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	log.Println(string(b))

	return nil
}
