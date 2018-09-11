package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const quickstatsEndpoint = "live/quickstats/v4/"

type QuickStatsData struct {
	Data *struct {
		Stats *QuickStats `json:"stats" bson:"stats"`
	} `json:"data" bson:"data"`
}

type QuickStats struct {
	Visits          int           `json:"visits" bson:"visits"`
	Links           int           `json:"links" bson:"links"`
	Direct          int           `json:"direct" bson:"direct"`
	Search          int           `json:"search" bson:"search"`
	Social          int           `json:"social" bson:"social"`
	Recirc          int           `json:"recirc" bson:"recirc"`
	Article         int           `json:"article" bson:"article"`
	Pages           int           `json:"pages" bson:"pages"`
	Platform        PlatformStats `bson:"platform" json:"platform"`
	PlatformEngaged PlatformStats `json:"platform_engaged" bson:"platform_engaged"`
	EngagedTime     ValueDistrib  `json:"engaged_time" bson:"engaged_time"`
	Loyalty         LoyaltyStats  `json:"loyalty" bson:"loyalty"`
	DOMLoadTime     ValueDistrib  `json:"domload" bson:"domload"`
}

type PlatformStats struct {
	M int `json:"m" bson:"m"`
	T int `json:"t" bson:"t"`
	D int `json:"d" bson:"d"`
	A int `json:"a" bson:"a"`
}

func (s *PlatformStats) Add(o PlatformStats) {
	s.M += o.M
	s.T += o.T
	s.D += o.D
	s.A += o.A
}

type LoyaltyStats struct {
	New       int `json:"new" bson:"new"`
	Loyal     int `json:"loyal" bson:"loyal"`
	Returning int `json:"returning" bson:"returning"`
}

func (cl *Client) FetchQuickStats(domain string) (*QuickStats, error) {
	var qs = url.Values{}
	qs.Set("all_platforms", "1")
	qs.Set("loyalty", "1")
	qs.Set("limit", "100")
	qs.Set("apikey", cl.APIKey)
	qs.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, quickstatsEndpoint, qs.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newHTTPCodeError(resp)
	}

	var r QuickStatsData
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, errMsgFailedToDecode)
	} else if r.Data == nil || r.Data.Stats == nil {
		return nil, ErrEmpty
	}

	return r.Data.Stats, nil
}
