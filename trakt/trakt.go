package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zm/traktshow/config"
)

const (
	apiURL       = "https://api.trakt.tv"
	clientID     = "275ab080baf9ab6044ddf8c5cb29463de65329b958df554631cf54f138125b6e"
	clientSecret = "d50ca71b9e4ad2651eb766e481299c156db771d4e873feb78742de29ca40e30d"
)

// Client struct holds the http client, credentials and token
type Client struct {
	client *http.Client
	config *config.Config
}

// NewClient creates a new Trakt client
func NewClient(cfg *config.Config) (*Client, error) {
	return &Client{
			client: &http.Client{},
			config: cfg,
		},
		nil
}

// Authenticate handles the device authentication flow
func (c *Client) Authenticate() (*config.Token, error) {
	// Step 1: Get device code
	deviceCode, err := c.getDeviceCode()
	if err != nil {
		return nil, fmt.Errorf("could not get device code: %w", err)
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
			return token, nil
		case <-time.After(time.Until(expiresAt)):
			return nil, fmt.Errorf("authentication timed out")
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
	req.Header.Set("trakt-api-key", clientID)

	if c.config.Token != nil {
		req.Header.Set("Authorization", "Bearer "+c.config.Token.AccessToken)
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

// GetHistory fetches user history for a specific page
func (c *Client) GetHistory(user string, limit, page int) ([]HistoryItem, error) {
	url := fmt.Sprintf("%s/users/%s/history?limit=%d&page=%d", apiURL, user, limit, page)
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var history []HistoryItem
	err = c.do(req, &history)
	return history, err
}

// GetAllHistory fetches all user history using pagination
func (c *Client) GetAllHistory(user string) ([]HistoryItem, error) {
	var allHistory []HistoryItem
	page := 1
	limit := 100 // Max limit per page

	for {
		fmt.Printf("Fetching page %d...\n", page)
		history, err := c.GetHistory(user, limit, page)
		if err != nil {
			return nil, err
		}

		if len(history) == 0 {
			// No more items, break the loop
			break
		}

		allHistory = append(allHistory, history...)
		page++
	}

	return allHistory, nil
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
	if err != nil {
		return nil, err
	}

	// For each watched item, fetch full show details to get accurate aired_episodes
	for i, item := range watched {
		if item.Show.IDs.Trakt != 0 {
			showDetails, err := c.GetShowDetails(item.Show.IDs.Trakt)
			if err != nil {
				fmt.Printf("Warning: Could not fetch details for show %s (Trakt ID: %d): %v\n", item.Show.Title, item.Show.IDs.Trakt, err)
				continue
			}
			watched[i].Show.AiredEpisodes = showDetails.AiredEpisodes
		}
	}

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

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func (c *Client) getDeviceCode() (*DeviceCodeResponse, error) {
	url := fmt.Sprintf("%s/oauth/device/code", apiURL)
	body := map[string]string{"client_id": clientID}
	req, err := c.newRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	var deviceCode DeviceCodeResponse
	err = c.do(req, &deviceCode)
	return &deviceCode, err
}

func (c *Client) pollForToken(deviceCode string) (*config.Token, error) {
	url := fmt.Sprintf("%s/oauth/device/token", apiURL)
	body := map[string]string{
		"code":          deviceCode,
		"client_id":     clientID,
		"client_secret": clientSecret,
	}
	req, err := c.newRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	var token config.Token
	if err := c.do(req, &token); err != nil {
		return nil, err
	}
	token.CreatedAt = time.Now().Unix()
	return &token, nil
}

// --- Data Structures ---

type HistoryItem struct {
	WatchedAt time.Time `json:"watched_at"`
	Show      struct {
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
		IDs           struct {
			Trakt int `json:"trakt"`
		} `json:"ids"`
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

// ShowDetails represents the detailed information of a show from Trakt API
type ShowDetails struct {
	IDs struct {
		Trakt int `json:"trakt"`
	} `json:"ids"`
	AiredEpisodes int `json:"aired_episodes"`
}

// GetShowDetails fetches detailed information for a show by its Trakt ID
func (c *Client) GetShowDetails(traktID int) (*ShowDetails, error) {
	url := fmt.Sprintf("%s/shows/%d?extended=full", apiURL, traktID)
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var showDetails ShowDetails
	err = c.do(req, &showDetails)
	if err != nil {
		return nil, err
	}
	return &showDetails, nil
}

