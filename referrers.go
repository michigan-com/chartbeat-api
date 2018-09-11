package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

var referrersEndpoint = "live/referrers/v3/"

type referrersResp struct {
	Referrers bson.M `bson:"referrers" json:"referrers"`
}

func (cl *Client) FetchReferrers(domain string) (map[string]int, error) {
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")
	queryParams.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, referrersEndpoint, queryParams.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newHTTPCodeError(resp)
	}

	var referrers referrersResp
	err = json.NewDecoder(resp.Body).Decode(&referrers)
	if err != nil {
		return nil, errors.Wrap(err, errMsgFailedToDecode)
	} else if len(referrers.Referrers) == 0 {
		return nil, ErrEmpty
	}

	result := make(map[string]int, len(referrers.Referrers))
	for k, v := range referrers.Referrers {
		c := int(v.(float64) + 0.5)
		result[k] = c
	}
	return result, nil
}
