package chartbeat

import (
	"errors"
)

const apiRoot = "http://api.chartbeat.com"

type Client struct {
	APIKey string
}

var ErrEmpty = errors.New("empty result")

const errMsgFailedToDecode = "incorrect response or parsing failure"
