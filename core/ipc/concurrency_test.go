package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestClientInferConcurrentWindows(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "blackice.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen unix: %v", err)
	}
	defer ln.Close()

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				var req map[string]any
				if err := json.NewDecoder(c).Decode(&req); err != nil {
					return
				}
				id, _ := req["window_id"].(float64)
				resp := map[string]any{"window_id": uint64(id), "anomaly_score": 0.42}
				b, _ := json.Marshal(resp)
				_, _ = c.Write(append(b, '\n'))
			}(conn)
		}
	}()

	client := NewClient(sock, 2*time.Second)
	const n = 100
	var wg sync.WaitGroup
	errCh := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		windowID := uint64(i + 1)
		go func(id uint64) {
			defer wg.Done()
			resp, err := client.Infer(context.Background(), InferenceRequest{WindowID: id})
			if err != nil {
				errCh <- err
				return
			}
			if resp.WindowID != id {
				errCh <- fmt.Errorf("window_id mismatch: got %d want %d", resp.WindowID, id)
			}
		}(windowID)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrency infer failed: %v", err)
		}
	}

	_ = ln.Close()
	<-serverDone
}
