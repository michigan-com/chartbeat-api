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
)

var topGeoEndpoint = "live/top_geo/v1"

type TopGeoSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Cities     []*TopGeo     `bson:"cities"`
}

type TopGeoResp struct {
	Geo TopGeo `bson:"geo:"`
}

type TopGeo struct {
	Source string `bson:"source"`
	Cities bson.M `bson:"cities"`
}

func (cl *Client) FetchTopGeo(domains []string) (*TopGeoSnapshot, error) {
	log.Info("Fetching topgeo...")
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")

	var wait sync.WaitGroup
	topGeoChannel := make(chan *TopGeo, len(domains))
	for _, domain := range domains {
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", apiRoot, topGeoEndpoint, queryParams.Encode())
		wait.Add(1)

		go func(url, domain string) {
			defer wait.Done()

			resp, err := fetchTopGeo(url)
			if err != nil {
				return
			}
			resp.Source = domain
			topGeoChannel <- resp
		}(url, domain)
	}
	wait.Wait()
	close(topGeoChannel)

	log.Info("...fetched topgeo")

	topGeo := make([]*TopGeo, 0, len(domains))
	for geo := range topGeoChannel {
		topGeo = append(topGeo, geo)
	}

	snapshot := &TopGeoSnapshot{}
	snapshot.Created_at = time.Now()
	snapshot.Cities = topGeo
	return snapshot, nil
}

func fetchTopGeo(url string) (*TopGeo, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to fetch Topgeo url %s\n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	topGeoResp := &TopGeoResp{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(topGeoResp)
	if err != nil {
		log.Errorf("\n\n\tFailed to decode Topgeo url %s\n\n\t", url, err)
		return nil, err
	}

	topGeo := &topGeoResp.Geo

	if len(topGeo.Cities) == 0 {
		err = errors.New(fmt.Sprintf("Cities array is empty for url %s", url))
		log.Errorf("\n\n\t%s\n\n\t", err)
		return nil, err
	}

	return topGeo, nil
}
