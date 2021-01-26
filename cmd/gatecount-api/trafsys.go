package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func getToken(baseURL, user, pass string) (string, error) {
	var vals = url.Values{
		"grant_type": {"password"},
		"username":   {user},
		"password":   {pass},
	}

	var c = &http.Client{}
	c.Timeout = time.Second * 15
	var resp, err = c.PostForm(baseURL+"/token", vals)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != 200 {
		l.Debugf("Received error from TrafSys.  Raw response body: %q", string(data))
		return "", fmt.Errorf("response code was %d", resp.StatusCode)
	}

	var serialized struct {
		Token string `json:"access_token"`
	}
	err = json.Unmarshal(data, &serialized)

	l.Debugf("Read token from TrafSys")

	return serialized.Token, err
}

func getTraffic(baseURL, token, start, end string) ([]*trafficCount, error) {
	var vals = url.Values{
		"SiteCode":                 {""},
		"DateFrom":                 {start},
		"DateTo":                   {end},
		"IncludeInternalLocations": {"false"},
		"DataSummedByDay":          {"false"},
	}

	var c = &http.Client{}
	c.Timeout = time.Second * 15
	var req, err = http.NewRequest("GET", baseURL+"/api/traffic?"+vals.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create GET request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	var resp *http.Response
	resp, err = c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != 200 {
		l.Debugf("Received error from TrafSys.  Raw response body: %q", string(data))
		return nil, fmt.Errorf("response code was %d", resp.StatusCode)
	}

	var counts []*trafficCount
	err = json.Unmarshal(data, &counts)
	l.Debugf("Read gate counts from TrafSys - parsing data for times and translation of gate ids")
	if err != nil {
		return nil, fmt.Errorf("couldn't parse JSON: %w", err)
	}

	for i, count := range counts {
		l.Debugf("Processing count #%d of %d: %s", i, len(counts), count)

		err = count.postProcess()
		if err != nil {
			return nil, fmt.Errorf("invalid count (#%d of %d): %w", i, len(counts), err)
		}

		l.Debugf("Post-processed count: %s", count)
	}

	return counts, nil
}
