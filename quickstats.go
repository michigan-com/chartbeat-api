package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const quickstatsEndpoint = "live/quickstats/v4"

type QuickStatsData struct {
	Data *struct {
		Stats *QuickStats `bson:"stats"`
	} `bson:"data"`
}

type QuickStats struct {
	Visits          int           `bson:"visits"`
	Links           int           `bson:"links"`
	Direct          int           `bson:"direct"`
	Search          int           `bson:"search"`
	Social          int           `bson:"social"`
	Recirc          int           `bson:"recirc"`
	Article         int           `bson:"article"`
	Platform        PlatformStats `json:"platform" bson:"platform"`
	PlatformEngaged PlatformStats `json:"platform_engaged" bson:"platform_engaged"`
	EngagedTime     ValueDistrib  `json:"engaged_time" bson:"engaged_time"`
	Loyalty         LoyaltyStats  `bson:"loyalty"`
	DOMLoadTime     ValueDistrib  `json:"domload"`
}

type PlatformStats struct {
	M int `bson:"m"`
	T int `bson:"t"`
	D int `bson:"d"`
	A int `bson:"a"`
}

func (s *PlatformStats) Add(o PlatformStats) {
	s.M += o.M
	s.T += o.T
	s.D += o.D
	s.A += o.A
}

type LoyaltyStats struct {
	New       int `bson:"new"`
	Loyal     int `bson:"loyal"`
	Returning int `bson:"returning"`
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
		return nil, errors.Errorf("HTTP error %v", resp.Status)
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
