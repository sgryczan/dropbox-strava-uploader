package dropbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"
)

const (
	port      = 8322
	tokenFile = "/app/dropbox_token.json"
)

var (
	oAuthConf      *oauth2.Config
	tokenSource    oauth2.TokenSource
	redirectURL    string
	clientID       string
	clientSecret   string
	callBackDomain string
	requiredVars   = []string{
		"DROPBOX_CLIENT_ID",
		"DROPBOX_CLIENT_SECRET",
		"CALLBACK_DOMAIN",
	}
	Debug bool
)

func init() {
	readVars()

	oAuthConf = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			"files.content.write",
			"files.metadata.write",
			"files.content.read",
		},
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://www.dropbox.com/oauth2/token",
			AuthURL:  "https://www.dropbox.com/oauth2/authorize",
		},
		RedirectURL: fmt.Sprintf("https://%s:%d/exchange_token", os.Getenv("CALLBACK_DOMAIN"), port),
	}

	t, err := readTokenFile()
	if err != nil {
		log.Println(err.Error())
		log.Printf("%+v", err)
	}

	client = oAuthConf.Client(context.Background(), t)

}

func readVars() {
	clientID = os.Getenv("DROPBOX_CLIENT_ID")
	clientSecret = os.Getenv("DROPBOX_CLIENT_SECRET")

	callBackDomain = os.Getenv("CALLBACK_DOMAIN")

	var missingRequiredVar bool
	for _, s := range requiredVars {
		v := os.Getenv(s)
		if v == "" {
			log.Printf("[dropbox] error: environment variable %s is empty. Please set a value and restart the program.\n", s)
			missingRequiredVar = true
		}
	}
	if missingRequiredVar {
		os.Exit(1)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// you should make this a template in your real application
	fmt.Fprintf(w, `<a href="%s">`, oAuthConf.AuthCodeURL("state", oauth2.SetAuthURLParam("token_access_type", "offline")))
	fmt.Fprint(w, `<img src="https://avatars.githubusercontent.com/u/559357?s=200&v=4" />`)
	fmt.Fprint(w, `</a>`)
}

func exchangeTokenHandler(w http.ResponseWriter, r *http.Request) {
	url, err := url.Parse(r.RequestURI)
	if err != nil {
		fmt.Fprintf(w, "error: %s", err.Error())
	}

	values := url.Query()
	code := values["code"][0]

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{Timeout: 2 * time.Second}
	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	tok, err := oAuthConf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	tokenSource = oAuthConf.TokenSource(ctx, tok)
	client = oAuthConf.Client(context.Background(), tok)
	err = WriteTokenFile()
	if err != nil {
		log.Print(err)
	}
	fmt.Fprintf(w, "%+v\n", tok)
}

func StartAuthServer() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.

	mux.HandleFunc("/exchange_token", exchangeTokenHandler)

	url := oAuthConf.AuthCodeURL("state", oauth2.SetAuthURLParam("token_access_type", "offline"))

	if !AuthIsGood() {
		fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
		fmt.Printf("No DropBox Auth Token Found. Visit http://%s:%d/ to get a token\n", os.Getenv("CALLBACK_DOMAIN"), port)
	}

	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

}

func readTokenFile() (*oauth2.Token, error) {
	t := &oauth2.Token{}
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}

	tokenSource = oAuthConf.TokenSource(context.Background(), t)
	return t, nil
}

func WriteTokenFile() error {
	t, err := tokenSource.Token()
	if err != nil {
		return err
	}

	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	err = os.WriteFile(tokenFile, b, 0644)
	if err != nil {
		return err
	}

	log.Printf("[dropbox] wrote auth token to %s\n", tokenFile)
	return nil
}
