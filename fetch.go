package main

import (
	"fmt"
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"

	"github.com/michigan-com/chartbeat-api/chartbeat"
	"github.com/michigan-com/chartbeat-api/parallel"
)

func fetch(originalSession *mgo.Session, chartb *chartbeat.Client, domains []string, gnapiDomain string) {
	q := parallel.New(10)

	q.Add(func() error {
		session := originalSession.Copy()
		defer session.Close()

		snapshot, err := chartb.FetchTopPages(domains)
		if err != nil {
			return err
		}

		return SaveTopPages(snapshot, session)
	})

	q.Add(func() error {
		session := originalSession.Copy()
		defer session.Close()

		snapshot, err := chartb.FetchQuickStats(domains)
		if err != nil {
			return err
		}

		return SaveQuickStats(snapshot, session)
	})

	q.Add(func() error {
		session := originalSession.Copy()
		defer session.Close()

		snapshot, err := chartb.FetchRecent(domains)
		if err != nil {
			return err
		}

		return SaveRecent(snapshot, session)
	})

	q.Add(func() error {
		session := originalSession.Copy()
		defer session.Close()

		snapshot, err := chartb.FetchReferrers(domains)
		if err != nil {
			return err
		}

		return SaveReferrers(snapshot, session)
	})

	q.Add(func() error {
		session := originalSession.Copy()
		defer session.Close()

		snapshot, err := chartb.FetchTopGeo(domains)
		if err != nil {
			return err
		}

		return SaveTopGeo(snapshot, session)
	})

	q.Add(func() error {
		session := originalSession.Copy()
		defer session.Close()

		snapshot, err := chartb.FetchTrafficSeries(domains)
		if err != nil {
			return err
		}

		return SaveTrafficSeries(snapshot, session)
	})

	err := q.Wait()
	if err != nil {
		log.Errorf("fetch failed: %v", err)
	}

	if gnapiDomain != "" {
		var wait sync.WaitGroup
		// Now hit all the necessary endpoints to update mapi
		// TODO do something different
		mapiUrls := []string{"popular", "quickstats", "topgeo", "referrers", "recent", "traffic-series"}
		for _, url := range mapiUrls {
			wait.Add(1)

			mapiUrl := fmt.Sprintf("%s/%s/", gnapiDomain, url)
			go func(mapiUrl string) {
				defer wait.Done()
				resp, err := http.Get(mapiUrl)
				if err != nil {
					return
				}
				defer resp.Body.Close()
			}(mapiUrl)
		}
		wait.Wait()
	} else {
		log.Info("No Gnapi domain specified, cannot update gnapi instance")
	}
}
