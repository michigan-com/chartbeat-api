package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var toppagesEndpoint = "live/toppages/v3"

type TopPagesData struct {
	Pages []*ArticleContent `json:"pages"`
}

type ArticleContent struct {
	Path     string        `json:"path"`
	Sections []string      `json:"sections"`
	Stats    *ArticleStats `json:"stats"`
	Title    string        `json:"title"`
	Authors  []string      `json:"authors"`
}

type ArticleStats struct {
	Visits          int           `json:"visits"`
	Loyalty         LoyaltyStats  `json:"loyalty"`
	Platform        PlatformStats `json:"platform"`
	PlatformEngaged PlatformStats `json:"platform_engaged"`
	Direct          int           `json:"direct"`
	Links           int           `json:"links"`
	Search          int           `json:"search"`
	Social          int           `json:"social"`
	Recirc          int           `json:"recirc"`
	Idle            int           `json:"idle"`
}

func (cl *Client) FetchTopPages(domain string) (*TopPagesData, error) {
	var queryParams = url.Values{}
	queryParams.Set("all_platforms", "1")
	queryParams.Set("loyalty", "1")
	queryParams.Set("limit", "100")
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, toppagesEndpoint, queryParams.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "toppages request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("toppages api call returned error %d", resp.StatusCode)
	}

	var t TopPagesData
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode toppages response")
	} else if len(t.Pages) == 0 {
		return nil, errors.New("Pages array length is zero")
	}

	return &t, nil
}
