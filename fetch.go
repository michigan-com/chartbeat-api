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

	f.fetchTopPages()
	f.fetchQuickStats()
	f.fetchRecent()
	f.fetchReferrers()
	f.fetchTopGeo()
	f.fetchTrafficSeries()

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

func (f *fetcher) fetchTopPages() {
	toppages := make(map[string]*chartbeat.TopPagesData, len(f.domains))

	g := f.netq.NewGroup()
	for _, d := range f.domains {
		domain := d
		g.Add(func() error {
			log.Infof("Fetching toppages from %s...", domain)
			result, err := f.chartb.FetchTopPages(domain)
			if result != nil {
				g.Sync(func() {
					toppages[domain] = result
				})
			}
			return err
		})
	}

	f.dbq.Add(func() error {
		log.Info("Waiting for toppages...")
		if !g.Wait() {
			log.Warning("Waiting for toppages done - FAILED")
			return nil
		}
		log.Info("Waiting for toppages done")
		return saveTopPages(toppages, f.db)
	})
}

func (f *fetcher) fetchQuickStats() {
	stats := make(map[string]*chartbeat.QuickStats, len(f.domains))

	g := f.netq.NewGroup()
	for _, d := range f.domains {
		domain := d
		g.Add(func() error {
			log.Infof("Fetching quickstats for %s...", domain)
			result, err := f.chartb.FetchQuickStats(domain)
			if result != nil {
				g.Sync(func() {
					stats[domain] = result
				})
			}
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

func (f *fetcher) fetchRecent() {
	recents := make(map[string]*chartbeat.RecentResp, len(f.domains))

	g := f.netq.NewGroup()
	for _, d := range f.domains {
		domain := d
		g.Add(func() error {
			log.Infof("Fetching recents from %s...", domain)
			result, err := f.chartb.FetchRecent(domain)
			if result != nil {
				g.Sync(func() {
					recents[domain] = result
				})
			}
			return err
		})
	}

	f.dbq.Add(func() error {
		log.Info("Waiting for recent...")
		if !g.Wait() {
			log.Warning("Waiting for recent done - FAILED!")
			return nil
		}

		log.Info("Waiting for recent done")
		return saveRecent(recents, f.db)
	})
}

func (f *fetcher) fetchReferrers() {
	referrers := make(map[string]*chartbeat.Referrers, len(f.domains))

	g := f.netq.NewGroup()
	for _, d := range f.domains {
		domain := d
		g.Add(func() error {
			log.Infof("Fetching referrers from %s...", domain)
			result, err := f.chartb.FetchReferrers(domain)
			if result != nil {
				g.Sync(func() {
					referrers[domain] = result
				})
			}
			return err
		})
	}

	f.dbq.Add(func() error {
		log.Info("Waiting for referrers...")
		if !g.Wait() {
			log.Warning("Waiting for referrers done - FAILED")
			return nil
		}
		log.Info("Waiting for referrers done")
		return saveReferrers(referrers, f.db)
	})
}

func (f *fetcher) fetchTopGeo() {
	topGeo := make(map[string]*chartbeat.TopGeoResp, len(f.domains))

	g := f.netq.NewGroup()
	for _, d := range f.domains {
		domain := d
		g.Add(func() error {
			log.Infof("Fetching top geo from %s...", domain)
			result, err := f.chartb.FetchTopGeo(domain)
			if result != nil {
				g.Sync(func() {
					topGeo[domain] = result
				})
			}
			return err
		})
	}

	f.dbq.Add(func() error {
		log.Info("waiting for top geo...")
		if !g.Wait() {
			log.Warning("Waiting for top geo done - FAILED")
			return nil
		}
		log.Info("Waiting for top geo done")
		return saveTopGeo(topGeo, f.db)
	})
}

func (f *fetcher) fetchTrafficSeries() {
	trafficSeries := make(map[string]*chartbeat.TrafficSeriesResp, len(f.domains))

	g := f.netq.NewGroup()
	for _, d := range f.domains {
		domain := d
		g.Add(func() error {
			log.Infof("Fetching traffic serires for %s...", domain)
			result, err := f.chartb.FetchTrafficSeries(domain)
			if result != nil {
				g.Sync(func() {
					trafficSeries[domain] = result
				})
			}
			return err
		})
	}

	f.dbq.Add(func() error {
		log.Info("Waiting for traffic series...")
		if !g.Wait() {
			log.Warning("Waiting for traffic series done - FAILED")
			return nil
		}
		log.Info("Waiting for traffic series done")
		return saveTrafficSeries(trafficSeries, f.db)
	})
}
