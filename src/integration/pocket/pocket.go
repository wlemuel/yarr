package pocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nkanaev/yarr/src/storage"
)

const (
	host         = "https://getpocket.com/v3"
	authorizeUrl = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"

	endpointAdd          = "/add"
	endpointRequestToken = "/oauth/request"
	endpointAuthorize    = "/oauth/authorize"

	// xErrorHeader used to parse error message from Headers on non-2XX responses
	xErrorHeader = "X-Error"

	defaultTimeout = 5 * time.Second

	// consumerKey = "107315-709b6b5789e8d9986798732"
)

type (
	requestTokenRequest struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectURI string `json:"redirect_uri"`
	}

	authorizeRequest struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}

	AuthorizeResponse struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}

	addRequest struct {
		URL         string `json:"url"`
		Title       string `json:"title,omitempty"`
		Tags        string `json:"tags,omitempty"`
		AccessToken string `json:"access_token"`
		ConsumerKey string `json:"consumer_key"`
	}

	// AddInput holds data necessary to create new item in Pocket list
	AddInput struct {
		URL         string
		Title       string
		Tags        []string
		AccessToken string
	}
)

func (i AddInput) validate() error {
	if i.URL == "" {
		return errors.New("required URL values is empty")
	}

	if i.AccessToken == "" {
		return errors.New("access token is empty")
	}

	return nil
}

func (i AddInput) generateRequest(consumerKey string) addRequest {
	return addRequest{
		URL:         i.URL,
		Tags:        strings.Join(i.Tags, ","),
		Title:       i.Title,
		AccessToken: i.AccessToken,
		ConsumerKey: consumerKey,
	}
}

// Client is a getpocket API client
type Client struct {
	client      *http.Client
	db          *storage.Storage
	consumerKey string
}

func NewClient(db *storage.Storage) *Client {
	accessToken = db.GetSettingsValue("access_token").(string)
	consumerKey = db.GetSettingsValue("consumer_key").(string)

	return &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		db:          db,
		consumerKey: consumerKey,
	}
}

// GetRequestToken obtains the request token that is used to authorize user in your application
func (c *Client) GetRequestToken(redirectUrl string) (string, error) {
	inp := &requestTokenRequest{
		ConsumerKey: c.consumerKey,
		RedirectURI: redirectUrl,
	}

	values, err := c.doHTTP(endpointRequestToken, inp)
	if err != nil {
		return "", err
	}

	if values.Get("code") == "" {
		return "", errors.New("empty request token in API response")
	}

	requestToken = values.Get("code")

	return requestToken, nil
}

// GetAuthorizationURL generates link to authorize user
func (c *Client) GetAuthorizationURL(redirectUrl string) (string, error) {
	if requestToken == "" || redirectUrl == "" {
		return "", errors.New("empty params")
	}

	return fmt.Sprintf(authorizeUrl, requestToken, redirectUrl), nil
}

// Authorize generates access token for user, that authorized in your app via link
func (c *Client) Authorize() (*AuthorizeResponse, error) {
	if requestToken == "" {
		return nil, errors.New("empty request token")
	}

	inp := &authorizeRequest{
		Code:        requestToken,
		ConsumerKey: c.consumerKey,
	}

	values, err := c.doHTTP(endpointAuthorize, inp)
	if err != nil {
		return nil, err
	}

	accessToken, username := values.Get("access_token"), values.Get("username")
	if accessToken == "" {
		return nil, errors.New("empty access token in API response")
	}

	settings := make(map[string]interface{})
	settings["access_token"] = accessToken

	c.db.UpdateSettings(settings)

	return &AuthorizeResponse{
		AccessToken: accessToken,
		Username:    username,
	}, nil
}

func (c *Client) add(input AddInput) (int64, error) {
	if err := input.validate(); err != nil {
		return 0, err
	}

	req := input.generateRequest(c.consumerKey)
	resp, err := c.doHTTP(endpointAdd, req)

	item := resp.Get("item")
	fmt.Printf("Pocket item: %s", item)

	return 1, err
}

// Add creates new item in Pocket list
func (c *Client) Add(url string) error {
	if accessToken == "" {
		return errors.New("failed to get access token")
	}

	_, err := c.add(AddInput{
		URL:         url,
		AccessToken: accessToken,
	})

	return err
}

func (c *Client) doHTTP(endpoint string, body interface{}) (url.Values, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return url.Values{}, errors.New("failed to marshal input body")
	}

	req, err := http.NewRequest(http.MethodPost, host+endpoint, bytes.NewBuffer(b))
	if err != nil {
		return url.Values{}, errors.New("failed to create new request")
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF8")

	resp, err := c.client.Do(req)
	if err != nil {
		return url.Values{}, errors.New("failed to send http request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("API Error: %s", resp.Header.Get(xErrorHeader))
		return url.Values{}, errors.New(err)
	}

	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return url.Values{}, errors.New("failed to read request body")
	}

	values, err := url.ParseQuery(string(respB))
	if err != nil {
		return url.Values{}, errors.New("failed to parse response body")
	}

	return values, nil
}

var (
	consumerKey  string = ""
	requestToken string = ""
	accessToken  string = ""
)
