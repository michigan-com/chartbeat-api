package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var topGeoEndpoint = "live/top_geo/v1"

type topGeoResp struct {
	Geo *TopGeo `json:"geo"`
}

type TopGeo struct {
	Cities map[string]int `json:"cities"`
}

func (cl *Client) FetchTopGeo(domain string) (*TopGeo, error) {
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")
	queryParams.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, topGeoEndpoint, queryParams.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newHTTPCodeError(resp)
	}

	var raw topGeoResp
	err = json.NewDecoder(resp.Body).Decode(&raw)
	if err != nil {
		return nil, errors.Wrap(err, errMsgFailedToDecode)
	} else if raw.Geo == nil || len(raw.Geo.Cities) == 0 {
		return nil, ErrEmpty
	}

	return raw.Geo, nil
}
