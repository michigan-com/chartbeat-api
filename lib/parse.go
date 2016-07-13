package lib

import (
	"net/url"
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
		if IsInvalidAuthor(testAuthor) {
			continue
		}
		authors = append(authors, testAuthor)
	}

	return authors
}

func IsInvalidAuthor(author string) bool {
	author = strings.ToLower(author)
	regex := regexp.MustCompile(`(and )?by `)
	author = regex.ReplaceAllString(author, "")

	if author == "" {
		return true
	}

	videoRegex := regexp.MustCompile(`^video\||^story pages gallery\|`)
	return videoRegex.MatchString(author)
}

/*
	Chartbeat queries have a GET parameter "host", which represents the host
 	we're getting data on. Pull the host from the url and return it.
	Return host (e.g. freep.com)
	Return "" if we don't find one
*/
func GetHostFromParams(inputUrl string) (string, error) {
	var host string
	var err error

	parsed, err := url.Parse(inputUrl)
	if err != nil {
		return host, err
	}

	hosts := parsed.Query()["host"]
	if len(hosts) > 0 {
		host = hosts[0]
	}

	return host, err
}

/*
	Chartbeat queries have a GET parameter "host", which represents the host
 	we're getting data on. Pull the host from the url and return it. Strip off the
	domain from it
	Return host (e.g. freep)
	Return "" if we don't find one
*/
func GetHostFromParamsAndStrip(inputUrl string) (string, error) {
	host, err := GetHostFromParams(inputUrl)
	host = strings.Replace(host, ".com", "", 1)
	return host, err
}

/*
	Get the host from a url, and strip it
*/
func GetDomainFromURL(inputUrl string) string {
	if !strings.HasPrefix(inputUrl, "http") {
		inputUrl = "http://" + inputUrl
	}

	parsedUrl, err := url.Parse(inputUrl)
	if err != nil {
		return ""
	}

	return parsedUrl.Host
}
