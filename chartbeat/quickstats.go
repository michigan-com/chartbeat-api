package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type QuickStats struct {
	Visits          int           `bson:"visits"`
	Links           int           `bson:"links"`
	Direct          int           `bson:"direct"`
	Search          int           `bson:"search"`
	Social          int           `bson:"social"`
	Recirc          int           `bson:"recirc"`
	Article         int           `bson:"article"`
	PlatformEngaged PlatformStats `json:"platform_engaged" bson:"platform_engaged"`
	Loyalty         LoyaltyStats  `bson:"loyalty"`
}

type PlatformStats struct {
	M int `bson:"m"`
	T int `bson:"t"`
	D int `bson:"d"`
	A int `bson:"a"`
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
		return nil, errors.Wrap(err, "quickstats request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("quickstats returned error %d", resp.StatusCode)
	}

	var r quickStatsResp
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode quickstats json")
	} else if r.Data == nil || r.Data.Stats == nil {
		return nil, errors.New("quickstats data or stats is nil")
	}

	return r.Data.Stats, nil
}

const quickstatsEndpoint = "live/quickstats/v4"

type quickStatsResp struct {
	Data *struct {
		Stats *QuickStats `bson:"stats"`
	} `bson:"data"`
}