package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WorkOrder struct {
	ID             string   `json:"id"`
	DispatchStatus string   `json:"dispatch_status"`
	Photos         []string `json:"photos"`
	FollowUp       string   `json:"follow_up"`
}

func snapshotKey(t time.Time) string {
	return fmt.Sprintf("snapshots/%s.json", t.UTC().Format("2006-01-02"))
}

func shouldUpload(orders []WorkOrder) bool { return len(orders) > 0 }

func uploadSnapshot(c *client, bucket string, now time.Time, orders []WorkOrder, put func(string, []byte) error) error {
	if !shouldUpload(orders) {
		return nil
	}
	b, err := json.Marshal(struct {
		GeneratedAt time.Time   `json:"generated_at"`
		WorkOrders  []WorkOrder `json:"work_orders"`
	}{now.UTC(), orders})
	if err != nil {
		return err
	}
	url, err := c.presign(bucket, snapshotKey(now))
	if err != nil {
		return err
	}
	if err := put(url, b); err != nil {
		return err
	}
	return nil
}

func httpPut(url string, body []byte) error {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("upload status %d", res.StatusCode)
	}
	return nil
}
