package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Record struct in the log
type Record struct {
	Value  []byte `json:"value"`  // Record value
	Offset uint64 `json:"offset"` // Position in log
}

// Log struct
type Log struct {
	mu      sync.Mutex
	records []Record // Slice to store records
}

// Log Function --> append new record to the Log
func (c *Log) Append(record Record) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	record.Offset = uint64(len(c.records))
	c.records = append(c.records, record)
	return record.Offset
}

// Log Function --> read a record in the Log with specific offset
func (c *Log) Read(offset uint64) (Record, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if offset >= uint64(len(c.records)) {
		return Record{}, http.ErrAbortHandler
	}
	return c.records[offset], nil
}

// Manage allowed methods
func (c *Log) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		c.handleProduce(w, r)
	case "GET":
		c.handleConsume(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Manage POST request
func (c *Log) handleProduce(w http.ResponseWriter, r *http.Request) {
	var record Record
	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	offset := c.Append(record)
	resp := map[string]uint64{"offset": offset}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Manage GET request
func (c *Log) handleConsume(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Offset uint64 `json:"offset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := c.Read(req.Offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(record); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	log := &Log{}
	http.ListenAndServe(":8080", log)
}
