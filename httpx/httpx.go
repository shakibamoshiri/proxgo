package httpx

import (
    stdhttp "net/http"
    "io"
)

func Delete(url string) (*stdhttp.Response, error) {
	req, err := stdhttp.NewRequest(stdhttp.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	return stdhttp.DefaultClient.Do(req)
}

func Put(url string, conType string, body io.Reader) (*stdhttp.Response, error) {
	req, err := stdhttp.NewRequest(stdhttp.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
    req.Header.Set("Content-Type", conType)
	return stdhttp.DefaultClient.Do(req)
}
