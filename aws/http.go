package aws

import "io"
import "net/http"

type IHttpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}
