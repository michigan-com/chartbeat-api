package lib

import (
	"testing"
)

var badAuthors = []string{
	"Video|Graham Couch / Lsj",
	"Video|Newslook",
	"Story Pages Gallery|Charlie Riedel",
	"Video|| Judy Putnam",
	"",
}

var goodAuthors = []string{
	"this is a good author",
	"test author",
	"asdfasdf",
	"John Smith",
}

type DomainTest struct {
	Url    string
	Domain string
}

func TestInvalidAuthor(t *testing.T) {
	for _, author := range badAuthors {
		if !IsInvalidAuthor(author) {
			t.Fatalf("Author '%s' should be marked as invalid", author)
		}
	}
	for _, author := range goodAuthors {
		if IsInvalidAuthor(author) {
			t.Fatalf("Author '%s' should be okay", author)
		}
	}
}

func TestParseAuthor(t *testing.T) {
	moreBadAuthors := append(badAuthors, []string{
		"and by ",
		"by ",
	}...)
	moreGoodAuthors := append(goodAuthors, []string{
		"and by ur mom",
		"by this one guy",
		"this one guy",
	}...)
	for _, author := range moreBadAuthors {
		authors := ParseAuthor(author)
		if len(authors) != 0 {
			t.Fatalf("Author '%s' is invalid, should not have returned any authors", author)
		}
	}
	for _, author := range moreGoodAuthors {
		authors := ParseAuthor(author)
		if len(authors) != 1 {
			t.Fatalf("Author '%s' should be only one author", author)
		}
	}
}

func TestGetDomainFromUrl(t *testing.T) {
	urlTest := []DomainTest{
		DomainTest{
			"http://google.com",
			"google.com",
		},
		DomainTest{
			"freep.com/story/news/local/michigan/detroit/2016/07/13/aclu-sues-wayne-co-treasurer-others-over-foreclosures/87025762/",
			"freep.com",
		},
		DomainTest{
			"http://freep.com/story/news/local/michigan/detroit/2016/07/13/aclu-sues-wayne-co-treasurer-others-over-foreclosures/87025762/",
			"freep.com",
		},
		DomainTest{
			"http://www.usatoday.com/story/news/politics/onpolitics/2016/07/13/scotus-ruth-bader-ginsburg-trump/87024248/",
			"www.usatoday.com",
		},
		DomainTest{
			"facebook.com",
			"facebook.com",
		},
	}

	for _, test := range urlTest {
		domain := GetDomainFromURL(test.Url)
		if domain != test.Domain {
			t.Fatalf("Domains dont match:\n\n\texpected: %s\n\tactual: %s", test.Domain, domain)
		}
	}
}
