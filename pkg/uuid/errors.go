package uuid

import (
	"fmt"
	"io"
)

// Shows the actual status code, as well as the response body.
// Shows the error instead if it can't read the response body.
type ErrHttpRequestFailed struct {
	StatusCode int
	Body       string
}

func NewErrHttpRequestFailed(status int, resBody io.ReadCloser) *ErrHttpRequestFailed {
	var body string
	b, err := io.ReadAll(resBody)
	if err != nil {
		body = err.Error()
	} else {
		body = string(b)
	}
	return &ErrHttpRequestFailed{
		StatusCode: status,
		Body:       body,
	}
}

func (e *ErrHttpRequestFailed) Error() string {
	return fmt.Sprintf("HTTP Request returned with Status Code %d, expected 200. Response body: %s", e.StatusCode, e.Body)
}
