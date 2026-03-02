package runtime

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"blackice/core/capture"
)

type fixedSource struct {
	packets []capture.Packet
}

func (f fixedSource) Run(ctx context.Context, out chan<- capture.Packet) {
	tick := time.NewTicker(2 * time.Millisecond)
	defer tick.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			out <- f.packets[i%len(f.packets)]
			i++
		}
	}
}

func TestServiceEscalatesToIPC(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "blackice.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	var calls atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				var req map[string]any
				_ = json.NewDecoder(c).Decode(&req)
				calls.Add(1)
				_, _ = c.Write([]byte(`{"window_id":1,"anomaly_score":0.95}` + "\n"))
			}(conn)
		}
	}()

	// low port entropy + high churn + high syn/ack should escalate
	s := NewService(Config{
		Window:           40 * time.Millisecond,
		PythonSocket:     sock,
		InferenceTimeout: 200 * time.Millisecond,
		Source: fixedSource{packets: []capture.Packet{
			{SrcIP: "10.0.1.1", DstPort: 80, Bytes: 70, SYN: true, ACK: false},
			{SrcIP: "10.0.1.2", DstPort: 80, Bytes: 70, SYN: true, ACK: false},
			{SrcIP: "10.0.1.3", DstPort: 80, Bytes: 70, SYN: true, ACK: false},
		}},
	})

	runCtx, runCancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer runCancel()
	_ = s.Run(runCtx)

	if calls.Load() == 0 {
		t.Fatal("expected at least one IPC inference call")
	}
}
