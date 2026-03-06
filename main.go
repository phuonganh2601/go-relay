package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type CacheItem struct {
	Data      []byte
	ExpiresAt time.Time
}

var cache = map[string]CacheItem{}
var mutex sync.RWMutex

const cacheTTL = 60 * time.Second

func getCache(key string) ([]byte, bool) {

	mutex.RLock()
	item, ok := cache[key]
	mutex.RUnlock()

	if !ok || time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Data, true
}

func setCache(key string, data []byte) {

	mutex.Lock()
	cache[key] = CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(cacheTTL),
	}
	mutex.Unlock()
}

func gzipResponse(w http.ResponseWriter, data []byte) {

	w.Header().Set("Content-Encoding", "gzip")

	gz := gzip.NewWriter(w)
	defer gz.Close()

	gz.Write(data)
}

func fetchHandler(w http.ResponseWriter, r *http.Request) {

	target := r.URL.Query().Get("url")

	if target == "" {
		http.Error(w, "missing url", 400)
		return
	}

	_, err := url.ParseRequestURI(target)
	if err != nil {
		http.Error(w, "invalid url", 400)
		return
	}

	// cache
	if data, ok := getCache(target); ok {

		if r.Header.Get("Accept-Encoding") == "gzip" {
			gzipResponse(w, data)
		} else {
			w.Write(data)
		}

		return
	}

	client := http.Client{
		Timeout: 20 * time.Second,
	}

	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	setCache(target, body)

	if r.Header.Get("Accept-Encoding") == "gzip" {

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		gz.Write(body)
		gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.Write(buf.Bytes())

	} else {

		w.Write(body)

	}
}

func main() {

	http.HandleFunc("/fetch", fetchHandler)

	log.Println("relay server running :8080")

	http.ListenAndServe(":8080", nil)
}