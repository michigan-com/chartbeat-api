package chartbeat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/lib"
)

var referrersEndpoint = "live/referrers/v3"

type ReferrersSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Referrers  []*Referrers  `bson:"referrers"`
}

type Referrers struct {
	Source    string `json:"source"`
	Referrers bson.M `bson:"referrers" json:"referrers"`
}

func (cl *Client) FetchReferrers(domains []string) (*ReferrersSnapshot, error) {
	log.Info("Fetching referrers...")
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")

	var wait sync.WaitGroup
	referrersChannel := make(chan *Referrers, len(domains))
	for _, domain := range domains {
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", apiRoot, referrersEndpoint, queryParams.Encode())
		wait.Add(1)

		go func(url string) {
			defer wait.Done()

			resp, err := fetchReferrers(url)
			if err != nil {
				return
			}
			referrersChannel <- resp
		}(url)
	}
	wait.Wait()
	close(referrersChannel)

	log.Info("...referrers fetched")

	referrers := make([]*Referrers, 0, len(domains))
	for referrer := range referrersChannel {
		referrers = append(referrers, referrer)
	}

	snapshot := &ReferrersSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Referrers = referrers
	return snapshot, nil
}

func fetchReferrers(url string) (*Referrers, error) {

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch referrers Url: %s:\n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	referrers := &Referrers{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(referrers)
	if err != nil {
		log.Errorf("\n\n\tFailed to decode referrers Url: %s\n\n\t%v\n", url, err)
		return nil, err
	}

	if len(referrers.Referrers) == 0 {
		err = errors.New(fmt.Sprintf("No referrers for url %s", url))
		log.Errorf("\n\n\t%s\n\n\t", err)
		return nil, err
	}

	referrers.Source, _ = lib.GetHostFromParamsAndStrip(url)

	return referrers, nil
}
