package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	URLPrefix = "envx+"

	fileScheme = "file"
)

var (
	errHTTPStatus = errors.New("http status error")

	Expansions = []ExpansionFunc{
		ExpandVars,
		ExpandURLs,
	}
	HTTPClient = &http.Client{
		Timeout: 3 * time.Second,
	}
)

type ExpansionFunc func(context.Context, map[string]string) (map[string]string, error)

func ExpandVars(ctx context.Context, m map[string]string) (map[string]string, error) {
	out := map[string]string{}

	for key, value := range m {
		out[key] = os.Expand(
			value,
			func(k string) string {
				if v, ok := m[k]; ok {
					return v
				}

				return ""
			},
		)
	}

	return out, nil
}

func ExpandURLs(ctx context.Context, m map[string]string) (map[string]string, error) {
	out := map[string]string{}

	for key, value := range m {
		if !strings.HasPrefix(value, URLPrefix) {
			out[key] = value
			continue
		}

		expanded, err := expandURL(ctx, value)
		if err != nil {
			return out, err
		}

		out[key] = expanded
	}

	return out, nil
}

func expandURL(ctx context.Context, val string) (string, error) {
	u, err := url.Parse(strings.TrimPrefix(val, URLPrefix))
	if err != nil {
		return val, err
	}

	if u.Scheme == fileScheme {
		b, err := os.ReadFile(strings.TrimPrefix(u.String(), fileScheme+"://"))
		if err != nil {
			return val, err
		}

		return string(b), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return val, err
	}

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return val, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return val, fmt.Errorf("%w: %v", errHTTPStatus, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return val, err
	}

	return string(respBody), nil
}
