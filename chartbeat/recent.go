package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var recentEndpoint = "live/recent/v3"

type RecentResp struct {
	Recents []Recent
}

type Recent struct {
	Lat      float32 `json:"lat"`
	Lng      float32 `json:"lng"`
	Title    string  `json:"title"`
	Url      string  `json:"path"`
	Host     string  `json:"domain"`
	Platform string  `json:"platform"`
}

func (cl *Client) FetchRecent(domain string) (*RecentResp, error) {
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")
	queryParams.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, recentEndpoint, queryParams.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "recent request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("recents returned error %d", resp.StatusCode)
	}

	recentArray := make([]Recent, 0, 100)
	err = json.NewDecoder(resp.Body).Decode(&recentArray)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode recent response")
	} else if len(recentArray) == 0 {
		return nil, errors.New("No recents returned")
	}

	var r = RecentResp{
		recentArray,
	}

	return &r, nil
}
