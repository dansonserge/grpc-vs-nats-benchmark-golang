package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"time"
)

type Result struct {
	Mode        string        `json:"mode"`
	Requests    int           `json:"requests"`
	Concurrency int           `json:"concurrency"`
	PayloadSize int           `json:"payload_size"`
	Durations   []int64       `json:"durations_ns"`
	TotalTime   time.Duration `json:"total_time"`
}

func percentile(sorted []int64, p float64) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted))*p + 0.5)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func analyze(r Result) {
	durs := append([]int64(nil), r.Durations...)
	sort.Slice(durs, func(i, j int) bool { return durs[i] < durs[j] })

	var sum int64
	for _, v := range durs {
		sum += v
	}
	avg := time.Duration(sum / int64(len(durs)))
	p50 := time.Duration(percentile(durs, 0.50))
	p95 := time.Duration(percentile(durs, 0.95))
	p99 := time.Duration(percentile(durs, 0.99))
	max := time.Duration(durs[len(durs)-1])
	ops := float64(r.Requests) / r.TotalTime.Seconds()

	fmt.Printf("%s benchmark:\n", r.Mode)
	fmt.Printf("  Requests: %d, Concurrency: %d, Payload: %d bytes\n", r.Requests, r.Concurrency, r.PayloadSize)
	fmt.Printf("  Avg: %v, P50: %v, P95: %v, P99: %v, Max: %v, Ops/sec: %.2f\n\n",
		avg, p50, p95, p99, max, ops)
}

func main() {
	dir := "results"

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read results directory: %v", err)
	}

	var grpcFile, natsFile string
	var grpcTime, natsTime int64

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}
		if filepath.Base(f.Name()) == "" {
			continue
		}

		if filepath.Base(f.Name())[0:4] == "grpc" || filepath.Base(f.Name())[len(f.Name())-9:] == "grpc.json" {
			if f.ModTime().Unix() > grpcTime {
				grpcTime = f.ModTime().Unix()
				grpcFile = filepath.Join(dir, f.Name())
			}
		}
		if filepath.Base(f.Name())[0:4] == "nats" || filepath.Base(f.Name())[len(f.Name())-9:] == "nats.json" {
			if f.ModTime().Unix() > natsTime {
				natsTime = f.ModTime().Unix()
				natsFile = filepath.Join(dir, f.Name())
			}
		}
	}

	if grpcFile != "" {
		data, err := ioutil.ReadFile(grpcFile)
		if err != nil {
			log.Fatalf("failed to read %s: %v", grpcFile, err)
		}
		var r Result
		if err := json.Unmarshal(data, &r); err != nil {
			log.Fatalf("failed to parse %s: %v", grpcFile, err)
		}
		analyze(r)
	}

	if natsFile != "" {
		data, err := ioutil.ReadFile(natsFile)
		if err != nil {
			log.Fatalf("failed to read %s: %v", natsFile, err)
		}
		var r Result
		if err := json.Unmarshal(data, &r); err != nil {
			log.Fatalf("failed to parse %s: %v", natsFile, err)
		}
		analyze(r)
	}
}
