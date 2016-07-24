package main

import (
	"encoding/json"
	"testing"
	"time"

	chartb "github.com/michigan-com/chartbeat-api/chartbeat"
)

func TestQuickstatsPipeline(t *testing.T) {
	testData := []byte(`{
		"data": {
			"stats": {
				"visits": 1,
				"links": 2,
				"direct": 3,
				"search": 4,
				"social": 5,
				"article": 6,
				"platform_engaged": {
					"m": 7,
					"t": 8,
					"d": 9,
					"a": 10
				},
				"loyalty": {
					"new": 11,
					"loyal": 12,
					"returning": 13
				}
			}
		}
	}`)

	data := chartb.QuickStatsData{}

	t.Log("it should unmarshal QuickStats json data into the QuickStatsData struct")

	err := json.Unmarshal(testData, &data)
	if err != nil {
		t.Errorf("JSON unmarshall failed, check QuickStatsData struct, %s", err)
	}

	now := time.Now()
	freep := map[string]*chartb.QuickStats{"freep.com": data.Data.Stats}
	formatData := formatQuickStats(now, freep)
	fData := formatData.Stats[0]

	t.Log("it should properly format the json data for our mongodb store")

	if fData.Visits != 1 {
		t.Errorf("Visits -- Expected 1, Actual %d", fData.Visits)
	}

	if fData.Links != 2 {
		t.Errorf("Links -- Expected 2, Actual %d", fData.Links)
	}

	if fData.Direct != 3 {
		t.Errorf("Direct -- Expected 3, Actual %d", fData.Direct)
	}

	if fData.Search != 4 {
		t.Errorf("Search -- Expected 4, Actual %d", fData.Search)
	}

	if fData.Social != 5 {
		t.Errorf("Social -- Expected 5, Actual %d", fData.Social)
	}

	if fData.Article != 6 {
		t.Errorf("Article -- Expected 6, Actual %d", fData.Article)
	}

	platformData := formatPlatformValues(now, freep)
	pData, ok := platformData["freep.com"].(PlatformStatsValue)
	if !ok {
		t.Errorf("formatPlatformValues should convert to PlatformStatsValue")
	}

	if pData.M != 7 {
		t.Errorf("M -- Expected 7, Actual %d", pData.M)
	}

	if pData.T != 8 {
		t.Errorf("T -- Expected 8, Actual %d", pData.T)
	}

	if pData.D != 9 {
		t.Errorf("D -- Expected 9, Actual %d", pData.D)
	}

	if pData.A != 10 {
		t.Errorf("A -- Expected 10, Actual %d", pData.A)
	}
}
