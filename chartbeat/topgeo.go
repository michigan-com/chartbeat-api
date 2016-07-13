package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

var topGeoEndpoint = "live/top_geo/v1"

type TopGeoResp struct {
	Geo *TopGeo `json:"geo"`
}

type TopGeo struct {
	Cities bson.M `json:"cities"`
}

func (cl *Client) FetchTopGeo(domain string) (*TopGeoResp, error) {
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")
	queryParams.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, topGeoEndpoint, queryParams.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "top geo http request failed")
	}
	defer resp.Body.Close()

	topGeoResp := TopGeoResp{}
	err = json.NewDecoder(resp.Body).Decode(&topGeoResp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode topgeo json")
	} else if topGeoResp.Geo == nil || len(topGeoResp.Geo.Cities) == 0 {
		return nil, errors.New("No cities found")
	}

	return &topGeoResp, nil
}
