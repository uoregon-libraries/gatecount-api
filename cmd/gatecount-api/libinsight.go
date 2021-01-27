package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

type postJSON struct {
	Date      string `json:"date"`
	GateID    int    `json:"gate_id"`
	GateStart int    `json:"gate_start"`
	GateEnd   int    `json:"gate_end"`
}

func postCounts(libinsightURL string, aggregated map[string]*trafficCount) {
	var retry int
	var badKeys []string

	// Organize keys we need to post so their order makes debugging a tiny bit
	// less awful, and we can group them for bulk operations
	var keys, next []string
	for key, _ := range aggregated {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for retry < 10 {
		var size = 500 >> retry
		if size < 10 {
			size = 10
		}
		l.Infof("Posting %d counts in batches of %d", len(keys), size)

		for len(keys) > 0 {
			if size > len(keys) {
				size = len(keys)
			}
			next, keys = keys[:size], keys[size:]

			var batch = make([]*trafficCount, len(next))
			for i, key := range next {
				batch[i] = aggregated[key]
			}

			l.Debugf("Batching and posting %d counts to LibInsight", len(batch))
			var err = postBatch(libinsightURL, batch)
			if err != nil {
				l.Warnf("Unable to send data to LibInsight: %s", err)
				badKeys = append(badKeys, next...)
			}
		}

		if len(badKeys) == 0 {
			l.Infof("Successfully posted all traffic counts")
			return
		}

		keys = badKeys
		badKeys = make([]string, 0)
		var delay = (time.Second * 5) << retry
		if delay > time.Minute*5 {
			delay = time.Minute * 5
		}
		l.Infof("Delaying %s to retry remaining %d keys in queue", delay, len(keys))
		time.Sleep(delay)
		retry++
	}

	l.Fatalf("Too many retries; aborting")
}

func postBatch(postURL string, counts []*trafficCount) error {
	var traffic []postJSON

	for _, count := range counts {
		traffic = append(traffic, postJSON{
			Date:      count.When.Format("2006-01-02 15:00"),
			GateID:    count.GateID,
			GateStart: 0,
			GateEnd:   count.Ins,
		})
	}

	var c = &http.Client{}
	c.Timeout = time.Second * 15

	var trafData, err = json.Marshal(traffic)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON data: %w", err)
	}

	var req *http.Request
	req, err = http.NewRequest("POST", postURL, bytes.NewBuffer(trafData))
	if err != nil {
		return fmt.Errorf("unable to create POST request: %w", err)
	}

	var resp *http.Response
	resp, err = c.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != 200 {
		l.Debugf("Received error from LibInsight.  Raw response body: %q", string(data))
		return fmt.Errorf("response code was %d", resp.StatusCode)
	}

	var serialized struct {
		Response int `json:"response"`
	}
	err = json.Unmarshal(data, &serialized)
	if err != nil {
		return err
	}

	if serialized.Response != 1 {
		return fmt.Errorf("Expected a JSON response of 1, but got %d (full body: %q)", serialized.Response, string(data))
	}

	return nil
}
