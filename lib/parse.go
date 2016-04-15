package lib

import (
	"regexp"
	"strconv"
	"strings"
)

/*
	Get an article ID from the url
	If we don't find one in the url, send back the md5 hash of the string
*/
func GetArticleId(url string) int {
	// Given an article url, get the ID from it
	r := regexp.MustCompile("/([0-9]+)/{0,1}$")
	match := r.FindStringSubmatch(url)

	if len(match) <= 1 {
		return -1
	}

	i, err := strconv.Atoi(match[1])
	if err != nil {
		return -1
	}

	return i
}

func IsBlacklisted(url string) bool {
	blacklist := []string{
		"/videos/",
		"/police-blotter/",
		"/interactives/",
		"facebook.com",
		"/errors/404",
	}

	for _, item := range blacklist {
		if strings.Contains(url, item) {
			return true
		}
	}

	return false
}

func ParseAuthors(authors []string) []string {
	parsedAuthors := make([]string, 0, len(authors))

	for _, author := range authors {
		parsedAuthors = append(parsedAuthors, ParseAuthor(author)...)
	}

	return parsedAuthors
}

func ParseAuthor(author string) []string {
	splitAuthors := strings.Split(author, " and ")
	authors := make([]string, 0, len(splitAuthors))

	for _, testAuthor := range splitAuthors {
		// Parse out "by ..." and "and by..."
		regex := regexp.MustCompile(`(and )?by `)
		testAuthor = regex.ReplaceAllString(testAuthor, "")

		if testAuthor == "" {
			continue
		}
		authors = append(authors, testAuthor)
	}

	return authors
}
