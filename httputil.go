package chartbeat

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode/utf8"
)

func newHTTPCodeError(resp *http.Response) error {
	b, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1024*1024*10))
	if err == nil && len(b) > 0 && utf8.Valid(b) {
		s := string(b)
		s = strings.Replace(s, "\n", " ", -1)
		s = strings.Replace(s, "\r", "", -1)
		return fmt.Errorf("HTTP error %v: %s", resp.Status, s)
	} else {
		return fmt.Errorf("HTTP error %v", resp.Status)
	}
}
