// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

// Config contains the client configuration.
type Config struct {
	// Base URL of the service (https://keys.sylabs.io is used if not supplied).
	BaseURL string
	// Auth token to include in the Authorization header of each request (if supplied).
	AuthToken string
	// User agent to include in each request (if supplied).
	UserAgent string
	// HTTPClient to use to make HTTP requests (if supplied).
	HTTPClient *http.Client
}

// DefaultConfig is a configuration that uses default values.
var DefaultConfig = &Config{}

// PageDetails includes pagination details.
type PageDetails struct {
	// Maximum number of results per page (server may ignore or return fewer).
	Size int
	// Token for next page (advanced with each request, empty for last page).
	Token string
}

// Client describes the client details.
type Client struct {
	// Base URL of the service.
	BaseURL *url.URL
	// Auth token to include in the Authorization header of each request (if supplied).
	AuthToken string
	// User agent to include in each request (if supplied).
	UserAgent string
	// HTTPClient to use to make HTTP requests.
	HTTPClient *http.Client
}

// normalizeURL normalizes the scheme of the supplied URL. If an unsupported scheme is provided, an
// error is returned.
func normalizeURL(u *url.URL) (*url.URL, error) {
	switch u.Scheme {
	case "http", "https":
		return u, nil
	case "hkp":
		// The HKP scheme is HTTP and implies port 11371.
		newURL := *u
		newURL.Scheme = "http"
		if u.Port() == "" {
			newURL.Host = net.JoinHostPort(u.Hostname(), "11371")
		}
		return &newURL, nil
	case "hkps":
		// The HKPS scheme is HTTPS and implies port 443.
		newURL := *u
		newURL.Scheme = "https"
		return &newURL, nil
	default:
		return nil, fmt.Errorf("unsupported protocol scheme %q", u.Scheme)
	}
}

const defaultBaseURL = "https://keys.sylabs.io"

// NewClient sets up a new Key Service client with the specified base URL and auth token.
func NewClient(cfg *Config) (c *Client, err error) {
	if cfg == nil {
		cfg = DefaultConfig
	}

	// Determine base URL
	bu := defaultBaseURL
	if cfg.BaseURL != "" {
		bu = cfg.BaseURL
	}
	baseURL, err := url.Parse(bu)
	if err != nil {
		return nil, err
	}
	baseURL, err = normalizeURL(baseURL)
	if err != nil {
		return nil, err
	}

	c = &Client{
		BaseURL:   baseURL,
		AuthToken: cfg.AuthToken,
		UserAgent: cfg.UserAgent,
	}

	// Set HTTP client
	if cfg.HTTPClient != nil {
		c.HTTPClient = cfg.HTTPClient
	} else {
		c.HTTPClient = http.DefaultClient
	}

	return c, nil
}

// newRequest returns a new Request given a method, path, query, and optional body.
func (c *Client) newRequest(method, path, rawQuery string, body io.Reader) (r *http.Request, err error) {
	u := c.BaseURL.ResolveReference(&url.URL{
		Path:     path,
		RawQuery: rawQuery,
	})

	r, err = http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if v := c.AuthToken; v != "" {
		r.Header.Set("Authorization", fmt.Sprintf("BEARER %s", v))
	}
	if v := c.UserAgent; v != "" {
		r.Header.Set("User-Agent", v)
	}

	return r, nil
}
