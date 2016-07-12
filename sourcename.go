package main

import (
	"strings"
)

func getSourceFromDomain(domain string) string {
	return strings.Replace(domain, ".com", "", 1)
}
