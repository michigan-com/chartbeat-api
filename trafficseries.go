package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

var trafficSeriesEndpoint = "historical/traffic/series/"

type TrafficSeries struct {
	Start     time.Time     `json:"start" bson:"start"`
	End       time.Time     `json:"end" bson:"end"`
	Frequency time.Duration `json:"frequency" bson:"frequency"`

	Values []TrafficSeriesValue `json:"values"`
}

type TrafficSeriesValue struct {
	Time   time.Time `json:"time"`
	People int       `json:"people"`
}

type trafficSeriesResp struct {
	Data map[string]interface{} `json:"data"`
}

func (cl *Client) FetchTrafficSeries(domain string) (*TrafficSeries, error) {
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")
	queryParams.Set("host", domain)
	url := fmt.Sprintf("%s/%s?%s", apiRoot, trafficSeriesEndpoint, queryParams.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newHTTPCodeError(resp)
	}

	var raw trafficSeriesResp
	err = json.NewDecoder(resp.Body).Decode(&raw)
	if err != nil {
		return nil, errors.Wrap(err, errMsgFailedToDecode)
	}

	rawStart, ok := raw.Data["start"].(float64)
	if !ok {
		return nil, errors.Errorf("missing or invalid start: %v", raw.Data["start"])
	}
	rawEnd, ok := raw.Data["end"].(float64)
	if !ok {
		return nil, errors.Errorf("missing or invalid end: %v", raw.Data["end"])
	}
	rawFreq, ok := raw.Data["frequency"].(float64)
	if !ok {
		return nil, errors.Errorf("missing or invalid frequency: %v", raw.Data["frequency"])
	}

	result := TrafficSeries{
		Start:     time.Unix(int64(rawStart), 0),
		End:       time.Unix(int64(rawEnd), 0),
		Frequency: time.Duration(rawFreq) * time.Second,
	}

	series := raw.Data[domain].(map[string]interface{})["series"].(map[string]interface{})
	people := series["people"].([]interface{})

	tm := result.Start
	for idx, rawp := range people {
		var p int
		if rawp == nil {
			p = 0
		} else {
			f, ok := rawp.(float64)
			if !ok {
				return nil, errors.Errorf("invalid value at index %v: %v", idx, rawp)
			}
			p = int(f)
		}

		result.Values = append(result.Values, TrafficSeriesValue{
			Time:   tm,
			People: p,
		})

		tm = tm.Add(result.Frequency)
	}

	// fmt.Printf("Series for %v = %+v\n", domain, result)

	return &result, nil
}
