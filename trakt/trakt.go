package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	apiURL        = "https://api.trakt.tv"
	tokenFilename = "token.json"
	credsFilename = "credentials.json"
)

// Client struct holds the http client, credentials and token
type Client struct {
	client *http.Client
	creds  *Credentials
	token  *Token
}

// Credentials stores the client id and secret
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// Token stores the oauth tokens
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

// NewClient creates a new Trakt client
func NewClient() (*Client, error) {
	creds, err := loadCredentials()
	if err != nil {
		return nil, err
	}

	token, _ := loadToken()

	return &Client{
		client: &http.Client{},
		creds:  creds,
		token:  token,
	}, nil
}

// Authenticate handles the device authentication flow
func (c *Client) Authenticate() error {
	// Step 1: Get device code
	deviceCode, err := c.getDeviceCode()
	if err != nil {
		return fmt.Errorf("could not get device code: %w", err)
	}

	fmt.Printf("Go to %s and enter the code: %s\n", deviceCode.VerificationURL, deviceCode.UserCode)

	// Step 2: Poll for token
	interval := time.Duration(deviceCode.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	expiresAt := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	for {
		select {
		case <-ticker.C:
			token, err := c.pollForToken(deviceCode.DeviceCode)
			if err != nil {
				// continue polling
				continue
			}
			c.token = token
			return c.saveToken()
		case <-time.After(time.Until(expiresAt)):
			return fmt.Errorf("authentication timed out")
		}
	}
}

func (c *Client) newRequest(method, url string, body interface{}) (*http.Request, error) {
	var buf []byte
	if body != nil {
		var err error
		buf, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")

	if c.token != nil {
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
		req.Header.Set("trakt-api-key", c.creds.ClientID)
	} else {
		req.Header.Set("trakt-api-key", c.creds.ClientID)
	}

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return nil
}

// GetHistory fetches user history
func (c *Client) GetHistory(user string, limit int) ([]HistoryItem, error) {
	url := fmt.Sprintf("%s/users/%s/history?limit=%d", apiURL, user, limit)
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var history []HistoryItem
	err = c.do(req, &history)
	return history, err
}

// GetWatched fetches user watched progress
func (c *Client) GetWatched(user string) ([]WatchedItem, error) {
	url := fmt.Sprintf("%s/users/%s/watched/shows", apiURL, user)
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var watched []WatchedItem
	err = c.do(req, &watched)
	return watched, err
}

// GetStats fetches user stats
func (c *Client) GetStats(user string) (Stats, error) {
	url := fmt.Sprintf("%s/users/%s/stats", apiURL, user)
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return Stats{}, err
	}

	var stats Stats
	err = c.do(req, &stats)
	return stats, err
}

// --- Helper Functions ---

func getConfigPath(filename string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "trakt", filename)
}

func loadCredentials() (*Credentials, error) {
	path := getConfigPath(credsFilename)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("credentials not found at %s, please run 'traktshow config'", path)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("could not parse credentials: %w", err)
	}
	return &creds, nil
}

func loadToken() (*Token, error) {
	path := getConfigPath(tokenFilename)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (c *Client) saveToken() error {
	path := getConfigPath(tokenFilename)
	data, err := json.MarshalIndent(c.token, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)
}

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func (c *Client) getDeviceCode() (*DeviceCodeResponse, error) {
	url := fmt.Sprintf("%s/oauth/device/code", apiURL)
	body := map[string]string{"client_id": c.creds.ClientID}
	req, err := c.newRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	var deviceCode DeviceCodeResponse
	err = c.do(req, &deviceCode)
	return &deviceCode, err
}

func (c *Client) pollForToken(deviceCode string) (*Token, error) {
	url := fmt.Sprintf("%s/oauth/device/token", apiURL)
	body := map[string]string{
		"code":          deviceCode,
		"client_id":     c.creds.ClientID,
		"client_secret": c.creds.ClientSecret,
	}
	req, err := c.newRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	var token Token
	if err := c.do(req, &token); err != nil {
		return nil, err
	}
	token.CreatedAt = time.Now().Unix()
	return &token, nil
}

// --- Data Structures ---

type HistoryItem struct {
	WatchedAt time.Time `json:"watched_at"`
	Show struct {
		Title string `json:"title"`
	} `json:"show"`
	Episode struct {
		Season int    `json:"season"`
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"episode"`
}

type WatchedItem struct {
	Plays int `json:"plays"`
	Show  struct {
		Title         string `json:"title"`
		AiredEpisodes int    `json:"aired_episodes"`
	} `json:"show"`
}

type Stats struct {
	Movies struct {
		Watched int `json:"watched"`
	} `json:"movies"`
	Shows struct {
		Watched int `json:"watched"`
	} `json:"shows"`
	Episodes struct {
		Watched int `json:"watched"`
	} `json:"episodes"`
}