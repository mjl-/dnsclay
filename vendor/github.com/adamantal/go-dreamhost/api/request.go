package api

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"
)

const (
	methodGET = "GET"
)

func newRequest(
	ctx context.Context,
	params interface{},
) (*http.Request, error) {
	url, err := getURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, methodGET, url, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func getURL(params interface{}) (string, error) {
	url, err := url.Parse(apiURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}

	queryString, err := query.Values(params)
	if err != nil {
		return "", errors.Wrap(err, "failed to extract values from parameters")
	}

	url.RawQuery = queryString.Encode()

	return url.String(), nil
}
