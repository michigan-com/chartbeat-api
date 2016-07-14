package main

import (
	"testing"
	"encoding/json"
	chartb "github.com/michigan-com/chartbeat-api/chartbeat"
)

func TestToppagesPipeline(t *testing.T) {
	testData := []byte(`{
		"pages": [{
			"title": "Test title",
			"path": "http://freep.com/news/12345",
			"authors": ["author1", "author2", "author3"],
			"sections": ["section1", "section2"],
			"stats": {
				"visits": 111,
				"direct": 222,
				"links": 333,
				"search": 444,
				"social": 555,
				"recirc": 666,
				"idle": 777
			},
			"loyalty": {},
			"platform": {},
			"platform_engaged": {}
		}]
	}`)

	data := chartb.TopPagesData{}

	t.Log("it should unmarshal Toppages json data into the TopPagesData struct")

	err := json.Unmarshal(testData, &data)
	if err != nil {
		t.Errorf("JSON unmarshall failed, check TopPagesData struct, %s", err)
	}

	formatData := formatTopPages(&data)
	fData := formatData[0]

	t.Log("it should properly format the json data for our mongodb store")

	if fData.ArticleID != 12345 {
		t.Errorf("ArticleID -- Expected 12345, Actual %d", fData.ArticleID)
	}

	if fData.Domain != "freep.com" {
		t.Errorf("Domain -- Expected freep.com, Actual: %s", fData.Domain)
	}

	if fData.URL != "http://freep.com/news/12345" {
		t.Errorf("URL -- Expected http://freep.com/news/12345, Actual: %s", fData.Domain)
	}

	if fData.Headline != "Test title" {
		t.Errorf("Headline -- Expected: %s, Actual: %s", "Test title", fData)
	}

	if fData.Authors[0] != "author1" || fData.Authors[1] != "author2" || fData.Authors[2] != "author3" {
		t.Errorf("Authors -- invalid []string values")
	}

	if fData.Sections[0] != "section1" || fData.Sections[1] != "section2" {
		t.Errorf("Sections -- invalid []string values")
	}

	if fData.Stats.Visits != 111 {
		t.Errorf("Visits -- Expected: 111, Actual: %d", fData.Stats.Visits)
	}

	if fData.Stats.Direct != 222 {
		t.Errorf("Direct -- Expected: 222, Actual: %d", fData.Stats.Direct)
	}

	if fData.Stats.Links != 333 {
		t.Errorf("Links -- Expected: 333, Actual: %d", fData.Stats.Links)
	}

	if fData.Stats.Search != 444 {
		t.Errorf("Search -- Expected: 444, Actual: %d", fData.Stats.Search)
	}

	if fData.Stats.Social != 555 {
		t.Errorf("Links -- Expected: 555, Actual: %d", fData.Stats.Social)
	}

	if fData.Stats.Recirc != 666 {
		t.Errorf("Recirc -- Expected: 666, Actual %d", fData.Stats.Recirc)
	}

	if fData.Stats.Idle != 777 {
		t.Errorf("Idle -- Expected: 777, Actual %d", fData.Stats.Idle)
	}
}
