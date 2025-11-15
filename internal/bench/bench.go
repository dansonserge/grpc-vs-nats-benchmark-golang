package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	pb "grpc-vs-nats-benchmark-golang/proto"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

type result struct {
	Mode        string        `json:"mode"`
	Requests    int           `json:"requests"`
	Concurrency int           `json:"concurrency"`
	PayloadSize int           `json:"payload_size"`
	Durations   []int64       `json:"durations_ns"`
	TotalTime   time.Duration `json:"total_time"`
}

func saveResults(r result, mode string) {
	_ = os.MkdirAll("results", 0755)
	name := fmt.Sprintf("results/%d-%s.json", time.Now().Unix(), mode)
	f, err := os.Create(name)
	if err != nil {
		log.Printf("save results create: %v", err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	_ = enc.Encode(r)
	log.Printf("results saved to %s", name)
}

func stats(durs []int64, total time.Duration, reqs int, conc int) {
	sort.Slice(durs, func(i, j int) bool { return durs[i] < durs[j] })
	var sum int64
	for _, v := range durs {
		sum += v
	}
	avg := time.Duration(sum / int64(len(durs)))
	p50 := time.Duration(durs[len(durs)/2])
	p95 := time.Duration(durs[int(float64(len(durs))*0.95)])
	p99 := time.Duration(durs[int(float64(len(durs))*0.99)])
	max := time.Duration(durs[len(durs)-1])
	opsec := float64(reqs) / total.Seconds()

	fmt.Printf("requests=%d concurrency=%d avg=%v p50=%v p95=%v p99=%v max=%v ops/sec=%.2f",
		reqs, conc, avg, p50, p95, p99, max, opsec)
}

// RunGRPC benchmark
type grpcClient interface {
	Ping(ctx context.Context, in *pb.PingRequest, opts ...grpc.CallOption) (*pb.PingResponse, error)
}

func RunGRPC(client grpcClient, payload []byte, requests int, concurrency int, timeout time.Duration) {
	ch := make(chan int64, requests)
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, concurrency)
	start := time.Now()

	for i := 0; i < requests; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			t0 := time.Now()
			_, err := client.Ping(ctx, &pb.PingRequest{Payload: payload})
			took := time.Since(t0).Nanoseconds()
			if err != nil {
				ch <- timeout.Nanoseconds()
			} else {
				ch <- took
			}
			<-sem
		}()
	}

	wg.Wait()
	total := time.Since(start)
	close(ch)

	durs := make([]int64, 0, requests)
	for v := range ch {
		durs = append(durs, v)
	}

	stats(durs, total, requests, concurrency)
	r := result{"grpc", requests, concurrency, len(payload), durs, total}
	saveResults(r, "grpc")
}

// RunNATS benchmark
func RunNATS(nc *nats.Conn, subject string, payload []byte, requests int, concurrency int, timeout time.Duration) {
	ch := make(chan int64, requests)
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, concurrency)
	start := time.Now()

	for i := 0; i < requests; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			t0 := time.Now()
			_, err := nc.Request(subject, payload, timeout)
			took := time.Since(t0).Nanoseconds()
			if err != nil {
				ch <- timeout.Nanoseconds()
			} else {
				ch <- took
			}
			<-sem
		}()
	}

	wg.Wait()
	total := time.Since(start)
	close(ch)

	durs := make([]int64, 0, requests)
	for v := range ch {
		durs = append(durs, v)
	}

	stats(durs, total, requests, concurrency)
	r := result{"nats", requests, concurrency, len(payload), durs, total}

	fmt.Printf("NATS benchmark result: %+v\n", r)
	saveResults(r, "nats")
}
