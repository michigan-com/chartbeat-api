package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var trafficSeriesEndpoint = "historical/traffic/series/"

func (cl *Client) FetchTrafficSeries(domain string) (*TrafficSeriesResp, error) {
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")
	queryParams.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, trafficSeriesEndpoint, queryParams.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get traffic series url")
	}
	defer resp.Body.Close()

	// They use variable keys, based on the host, for each return
	// Decode, once for the data with static keys, and one for the data with variable keys
	trafficSeriesResp := TrafficSeriesResp{}
	err = json.NewDecoder(resp.Body).Decode(&trafficSeriesResp)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode traffic series resp")
	}

	return &trafficSeriesResp, nil
}

// TODO pull data out of this
type TrafficSeriesResp struct {
	Data map[string]interface{} `json:"data"`
}

type TrafficSeries struct {
	Series *struct {
		People []int `json:"people"`
	} `json:"series"`
}
