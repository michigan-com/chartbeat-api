package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

var referrersEndpoint = "live/referrers/v3"

type Referrers struct {
	Referrers bson.M `bson:"referrers" json:"referrers"`
}

func (cl *Client) FetchReferrers(domain string) (*Referrers, error) {
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")
	queryParams.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, referrersEndpoint, queryParams.Encode())

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch referrers Url: %s:\n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	referrers := Referrers{}
	err = json.NewDecoder(resp.Body).Decode(&referrers)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode referrers url")
	} else if len(referrers.Referrers) == 0 {
		return nil, errors.New("No referrers returned")
	}

	return &referrers, nil
}
