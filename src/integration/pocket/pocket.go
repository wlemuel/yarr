package pocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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
	accessToken = db.GetSettingsValueString("access_token")
	consumerKey = db.GetSettingsValueString("consumer_key")

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

	if requestToken, ok := values["code"].(string); ok {
		return requestToken, nil
	}

	return "", errors.New("empty request token in API response")
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

	if token, ok := values["access_token"].(string); ok {
		settings := make(map[string]interface{})
		settings["access_token"] = token

		c.db.UpdateSettings(settings)

		accessToken = token
	} else {
		return nil, errors.New("empty access token in API response")
	}

	var username string
	if user, ok := values["username"].(string); ok {
		username = user
	}

	return &AuthorizeResponse{
		AccessToken: accessToken,
		Username:    username,
	}, nil
}

func (c *Client) add(input AddInput) (string, error) {
	if err := input.validate(); err != nil {
		return "", err
	}

	req := input.generateRequest(c.consumerKey)
	resp, err := c.doHTTP(endpointAdd, req)
	if err != nil {
		return "", err
	}

	fmt.Printf("pocket add: %s, %v\n", input.URL, resp)

	if item, ok := resp["item"].(map[string]interface{}); ok {
		if item_id, ok := item["item_id"].(string); ok {
			return item_id, err
		}
	}

	return "", err
}

// Add creates new item in Pocket list
func (c *Client) Add(url string) (string, error) {
	accessToken = c.db.GetSettingsValueString("access_token")
	if accessToken == "" {
		return "", errors.New("failed to get access token")
	}

	item_id, err := c.add(AddInput{
		URL:         url,
		AccessToken: accessToken,
	})

	return item_id, err
}

func (c *Client) doHTTP(endpoint string, body interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, errors.New("failed to marshal input body")
	}

	req, err := http.NewRequest(http.MethodPost, host+endpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.New("failed to create new request")
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF8")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.New("failed to send http request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("API Error: %s", resp.Header.Get(xErrorHeader))
		return nil, errors.New(err)
	}

	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read request body")
	}

	var values map[string]interface{}
	err = json.Unmarshal(respB, &values)
	if err != nil {
		return nil, errors.New("failed to parse response body")
	}

	return values, nil
}

var (
	consumerKey  string = ""
	requestToken string = ""
	accessToken  string = ""
)
