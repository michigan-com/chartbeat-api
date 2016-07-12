package chartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"

	"github.com/michigan-com/chartbeat-api/lib"
)

var trafficSeriesEndpoint = "historical/traffic/series/"

func (cl *Client) FetchTrafficSeries(domains []string) (*TrafficSeriesSnapshot, error) {
	log.Info("Fetching traffic series...")
	queryParams := url.Values{}
	queryParams.Set("apikey", cl.APIKey)
	queryParams.Set("limit", "100")

	var start int
	var end int
	var frequency int
	var wait sync.WaitGroup
	trafficSeriesChannel := make(chan *Traffic, len(domains))
	for _, domain := range domains {
		queryParams.Set("host", domain)
		url := fmt.Sprintf("%s/%s?%s", apiRoot, trafficSeriesEndpoint, queryParams.Encode())
		wait.Add(1)
		go func(url string) {
			defer wait.Done()
			resp, err := fetchTrafficSeries(url)
			if err != nil {
				return
			}

			traffic := &Traffic{}
			traffic.Source = resp.Data.Source

			series := resp.GetSeries()
			if series != nil {
				traffic.Visits = series.Series.People
			}

			start = resp.Data.Start
			end = resp.Data.End
			frequency = resp.Data.Frequency

			trafficSeriesChannel <- traffic
		}(url)
	}
	wait.Wait()
	close(trafficSeriesChannel)

	log.Info("...fetched traffic series")

	trafficSeriesArray := make([]*Traffic, 0, len(domains))
	for traffic := range trafficSeriesChannel {
		trafficSeriesArray = append(trafficSeriesArray, traffic)
	}

	snapshot := &TrafficSeriesSnapshot{}
	snapshot.Start = start
	snapshot.End = end
	snapshot.Frequency = frequency
	snapshot.Traffic = trafficSeriesArray
	snapshot.Created_at = time.Now()
	return snapshot, nil
}

func fetchTrafficSeries(url string) (*TrafficSeriesIn, error) {

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("\n\n\tFailed to get traffic series url %s: \n\n\t%v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	// They use variable keys, based on the host, for each return
	// Decode, once for the data with static keys, and one for the data with variable keys
	trafficSeriesIn := &TrafficSeriesIn{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(trafficSeriesIn)
	if err != nil {
		log.Errorf("\n\n\tFailed to decode url (using TrafficSeriesIn): %s\n\n\t%v\n", url, err)
		return nil, err
	}

	trafficSeriesIn.Data.Source, _ = lib.GetHostFromParamsAndStrip(url)

	return trafficSeriesIn, nil
}

type TrafficSeriesSnapshot struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Created_at time.Time     `bson:"created_at"`
	Start      int           `bson:"start"`
	End        int           `bson:"end"`
	Frequency  int           `bson:"frequency"`
	Traffic    []*Traffic    `bson:"sites"`
}

type Traffic struct {
	Source string `bson:"source"`
	Visits []int  `bson:"visits"`
}

type TrafficSeriesIn struct {
	Data struct {
		Start     int    `json:"start"`
		End       int    `json:"end"`
		Frequency int    `json:"frequency"`
		Source    string `bson:"source"`

		Freep       *TrafficSeries `json:"freep.com"`
		DetroitNews *TrafficSeries `json:"detroitnews.com"`
		BattleCreek *TrafficSeries `json:"battlecreekenquirer.com"`
		Hometown    *TrafficSeries `json:"hometownlife.com"`
		Lansing     *TrafficSeries `json:"lansingstatejournal.com"`
		Livingston  *TrafficSeries `json:"livingstondaily.com"`
		Herald      *TrafficSeries `json:"thetimesherald.com"`

		// Usat
		UsaToday *TrafficSeries `json:"usatoday.com"`

		// Tennessean
		Tennessean *TrafficSeries `json:"tennessean.com"`

		// Central Ohio omg why are there so many sites help
		Mansfield        *TrafficSeries `json:"mansfieldnewsjournal.com"`
		Newark           *TrafficSeries `json:"newarkadvocate.com"`
		Zanesville       *TrafficSeries `json:"zanesvilletimesrecorder.com"`
		Chillicothe      *TrafficSeries `json:"chillicothegazette.com"`
		Lancaster        *TrafficSeries `json:"lancastereaglegazette.com"`
		Marion           *TrafficSeries `json:"marionstar.com"`
		TheNewsMessenger *TrafficSeries `json:"thenews-messenger.com"`
		Coshocton        *TrafficSeries `json:"coshoctontribune.com"`
		Bucyrus          *TrafficSeries `json:"bucyrustelegraphforum.com"`
		PortClinton      *TrafficSeries `json:"portclintonnewsherald.com"`

		// Central Ohio omg why are there so many sites help
		DesMoines    *TrafficSeries `json:"desmoinesregister.com"`
		PressCitizen *TrafficSeries `json:"press-citizen.com"`
		Juice        *TrafficSeries `json:"dmjuice.com"`
		HawkCentral  *TrafficSeries `json:"hawkcentral.com"`
	} `json:"data"`
}

func (h *TrafficSeriesIn) GetSeries() *TrafficSeries {
	if h.Data.Freep != nil {
		return h.Data.Freep
	} else if h.Data.DetroitNews != nil {
		return h.Data.DetroitNews
	} else if h.Data.BattleCreek != nil {
		return h.Data.BattleCreek
	} else if h.Data.Hometown != nil {
		return h.Data.Hometown
	} else if h.Data.Lansing != nil {
		return h.Data.Lansing
	} else if h.Data.Livingston != nil {
		return h.Data.Livingston
	} else if h.Data.Herald != nil {
		return h.Data.Herald
	} else if h.Data.UsaToday != nil {
		return h.Data.UsaToday
	} else if h.Data.Tennessean != nil {
		return h.Data.Tennessean
	} else if h.Data.Mansfield != nil {
		return h.Data.Mansfield
	} else if h.Data.Newark != nil {
		return h.Data.Newark
	} else if h.Data.Zanesville != nil {
		return h.Data.Zanesville
	} else if h.Data.Chillicothe != nil {
		return h.Data.Chillicothe
	} else if h.Data.Lancaster != nil {
		return h.Data.Lancaster
	} else if h.Data.Marion != nil {
		return h.Data.Marion
	} else if h.Data.TheNewsMessenger != nil {
		return h.Data.TheNewsMessenger
	} else if h.Data.Coshocton != nil {
		return h.Data.Coshocton
	} else if h.Data.Bucyrus != nil {
		return h.Data.Bucyrus
	} else if h.Data.PortClinton != nil {
		return h.Data.PortClinton
	} else if h.Data.DesMoines != nil {
		return h.Data.DesMoines
	} else if h.Data.PressCitizen != nil {
		return h.Data.PressCitizen
	} else if h.Data.Juice != nil {
		return h.Data.Juice
	} else if h.Data.HawkCentral != nil {
		return h.Data.HawkCentral
	}
	return nil
}

type TrafficSeries struct {
	Series *struct {
		People []int `json:"people"`
	} `json:"series"`
}
