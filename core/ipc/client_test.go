package ipc

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClientInferOverUnixSocket(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "blackice.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen unix: %v", err)
	}
	defer ln.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		var req map[string]any
		if err := json.NewDecoder(conn).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if req["window_id"].(float64) != 42 {
			t.Errorf("window_id mismatch got %v", req["window_id"])
		}
		_, _ = conn.Write([]byte(`{"window_id":42,"anomaly_score":0.91}` + "\n"))
	}()

	client := NewClient(sock, time.Second)
	resp, err := client.Infer(context.Background(), InferenceRequest{WindowID: 42})
	if err != nil {
		t.Fatalf("infer failed: %v", err)
	}
	if resp.WindowID != 42 || resp.AnomalyScore != 0.91 {
		t.Fatalf("unexpected response: %#v", resp)
	}
	<-done
}

func TestClientInferConnectionFailure(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.sock")
	_ = os.Remove(missing)
	client := NewClient(missing, 100*time.Millisecond)
	_, err := client.Infer(context.Background(), InferenceRequest{WindowID: 1})
	if err == nil {
		t.Fatal("expected connection failure")
	}
}
