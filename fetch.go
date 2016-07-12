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

type fetcher struct {
	netq    *parallel.Queue
	dbq     *parallel.Queue
	db      *mgo.Database
	chartb  *chartbeat.Client
	domains []string
}

func fetch(session *mgo.Session, chartb *chartbeat.Client, domains []string, gnapiDomain string) {
	f := &fetcher{
		netq:    parallel.New(10, "netq"),
		dbq:     parallel.New(1, "dbq"),
		db:      session.DB(""),
		chartb:  chartb,
		domains: domains,
	}

	f.netq.Add(func() error {
		snapshot, err := chartb.FetchTopPages(domains)
		if err != nil {
			return err
		}

		f.dbq.Add(func() error {
			return saveTopPages(snapshot, session)
		})
		return nil
	})

	f.fetchQuickStats()

	f.netq.Add(func() error {
		snapshot, err := chartb.FetchRecent(domains)
		if err != nil {
			return err
		}

		f.dbq.Add(func() error {
			return saveRecent(snapshot, session)
		})
		return nil
	})

	f.netq.Add(func() error {
		snapshot, err := chartb.FetchReferrers(domains)
		if err != nil {
			return err
		}

		f.dbq.Add(func() error {
			return saveReferrers(snapshot, session)
		})
		return nil
	})

	f.netq.Add(func() error {
		snapshot, err := chartb.FetchTopGeo(domains)
		if err != nil {
			return err
		}

		f.dbq.Add(func() error {
			return saveTopGeo(snapshot, session)
		})
		return nil
	})

	f.netq.Add(func() error {
		snapshot, err := chartb.FetchTrafficSeries(domains)
		if err != nil {
			return err
		}

		f.dbq.Add(func() error {
			return saveTrafficSeries(snapshot, session)
		})
		return nil
	})

	neterr := f.netq.Wait()
	dberr := f.dbq.Wait()

	if neterr != nil {
		log.Errorf("fetch failed: %v", neterr)
	}
	if dberr != nil {
		log.Errorf("saving failed: %v", dberr)
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

func (f *fetcher) fetchQuickStats() {
	stats := make(map[string]*chartbeat.QuickStats, len(f.domains))

	g := f.netq.NewGroup()
	for _, d := range f.domains {
		domain := d
		g.Add(func() error {
			log.Infof("Fetching quickstats for %s...", domain)
			result, err := f.chartb.FetchQuickStats(domain)
			g.Sync(func() {
				stats[domain] = result
			})
			return err
		})
	}

	f.dbq.Add(func() error {
		log.Info("Waiting for quickstats...")
		if !g.Wait() {
			log.Warning("Waiting for quickstats done - FAILED!")
			return nil
		}
		log.Info("Waiting for quickstats done")

		return saveQuickStats(stats, f.db)
	})
}
