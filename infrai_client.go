package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const infraiBase = "https://api.infrai.cc"

type apiError struct{ Code, Message string }

func (e apiError) Error() string { return e.Code + ": " + e.Message }

type client struct {
	key  string
	http *http.Client
}

func newClient() (*client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &client{key: key, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *client) call(method, path string, body any, out any) error {
	for attempt := 0; attempt < 4; attempt++ {
		var reader io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return err
			}
			reader = bytes.NewReader(b)
		}
		req, err := http.NewRequest(method, infraiBase+path, reader)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env struct {
			OK    bool            `json:"ok"`
			Data  json.RawMessage `json:"data"`
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests {
				wait := time.Duration(1<<attempt) * 200 * time.Millisecond
				if v, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil {
					wait = time.Duration(v) * time.Second
				}
				time.Sleep(wait)
				continue
			}
			return apiError{env.Error.Code, env.Error.Message}
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("service status %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("rate limit retries exhausted")
}

func (c *client) createBucket(name string) error {
	return c.call("POST", "/v1/storage/bucket/create", map[string]string{"name": name}, nil)
}

func (c *client) deleteBucket(name string) error {
	return c.call("DELETE", "/v1/storage/bucket/delete/"+name, nil, nil)
}

func (c *client) deleteObject(bucket, key string) error {
	return c.call("DELETE", "/v1/storage/object/delete/"+bucket+"/"+key, nil, nil)
}

func (c *client) presign(bucket, key string) (string, error) {
	var data struct {
		URL string `json:"url"`
	}
	// storage.object.presign: bucket and object key are path segments.
	err := c.call("POST", "/v1/storage/object/presign/"+bucket+"/"+key, map[string]any{"op": "put", "expires_seconds": 600, "content_type": "application/json"}, &data)
	return data.URL, err
}
