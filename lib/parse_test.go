package lib

import (
	"testing"
)

var badAuthors = []string{
	"Video|Graham Couch / Lsj",
	"Video|Newslook",
	"Pages Gallery|Charlie Riedel",
	"Video|| Judy Putnam",
	"",
}

var goodAuthors = []string{
	"this is a good author",
	"test author",
	"asdfasdf",
	"John Smith",
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
