package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

var (
	client       = &http.Client{}
	refreshMutex sync.Mutex
)

func httpPostJSON(url string, body any, withAuth bool) (*http.Response, error) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if withAuth {
		if at := loadAccessToken(); at != "" {
			req.Header.Set("Authorization", "Bearer "+at)
		}
	}
	return doRequest(req)
}

func httpGetJSON(url string, withAuth bool) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")
	if withAuth {
		if at := loadAccessToken(); at != "" {
			req.Header.Set("Authorization", "Bearer "+at)
		}
	}
	return doRequest(req)
}

func doRequest(req *http.Request) (*http.Response, error) {
	if at := loadAccessToken(); at != "" {
		req.Header.Set("Authorization", "Bearer "+at)
	}
	resp, err := client.Do(req)
	if err != nil { return nil, err }
	if resp.StatusCode != http.StatusUnauthorized { return resp, nil }
	// 401 自动刷新
	resp.Body.Close()
	refreshMutex.Lock()
	defer refreshMutex.Unlock()
	if err := refreshAccessToken(); err != nil { return nil, err }
	// retry once
	req2 := req.Clone(req.Context())
	if at := loadAccessToken(); at != "" { req2.Header.Set("Authorization", "Bearer "+at) }
	return client.Do(req2)
}

func refreshAccessToken() error {
	rt := loadRefreshToken()
	if rt == "" { return errors.New("no refresh token") }
	b, _ := json.Marshal(map[string]string{"refresh_token": rt})
	req, _ := http.NewRequest(http.MethodPost, apiURL("auth.refresh_token"), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	var body apiResponse[loginResp]
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { return err }
	if !body.Success { return errors.New(body.Message) }
	return saveAccessToken(body.Data.AccessToken)
}
