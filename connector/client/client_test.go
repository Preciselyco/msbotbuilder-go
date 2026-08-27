// Copyright (c) 2020 InfraCloud Technologies
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infracloudio/msbotbuilder-go/connector/auth"
	"github.com/stretchr/testify/assert"
)

type erroringRoundTripper struct {
	err error
}

func (rt erroringRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, rt.err
}

func newTestClient(t *testing.T, authURL string, authClient *http.Client) *ConnectorClient {
	t.Helper()

	config, err := NewClientConfig(auth.SimpleCredentialProvider{AppID: "id", Password: "secret"}, authURL)
	assert.NoError(t, err)
	config.AuthClient = authClient

	c, err := NewClient(config)
	assert.NoError(t, err)

	return c.(*ConnectorClient)
}

func TestGetToken_AuthClientDoError_ReturnsErrorWithoutPanic(t *testing.T) {
	authClient := &http.Client{
		Transport: erroringRoundTripper{err: errors.New("connection refused")},
	}
	client := newTestClient(t, "https://login.example.com/token", authClient)

	assert.NotPanics(t, func() {
		_, err := client.getToken(context.Background())
		assert.Error(t, err)
	})
}

func TestGetToken_Success_ReturnsAndCachesToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"Bearer","expires_in":3600,"access_token":"test-token"}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, srv.Client())

	token, err := client.getToken(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-token", token)

	assert.False(t, client.AuthCache.IsExpired())
	assert.Equal(t, "test-token", client.AuthCache.Keys)
}
