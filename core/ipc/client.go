package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"blackice/core/fastpath"
)

type InferenceRequest = fastpath.WindowFeatures

type InferenceResponse struct {
	WindowID     uint64  `json:"window_id"`
	AnomalyScore float64 `json:"anomaly_score"`
}

type Client struct {
	socketPath string
	timeout    time.Duration
}

func NewClient(socketPath string, timeout time.Duration) *Client {
	return &Client{socketPath: socketPath, timeout: timeout}
}

func (c *Client) Infer(ctx context.Context, req InferenceRequest) (InferenceResponse, error) {
	d := net.Dialer{Timeout: c.timeout}
	conn, err := d.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return InferenceResponse{}, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(c.timeout))
	enc := json.NewEncoder(conn)
	if err := enc.Encode(req); err != nil {
		return InferenceResponse{}, fmt.Errorf("encode request: %w", err)
	}

	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return InferenceResponse{}, fmt.Errorf("read response: %w", err)
	}
	var resp InferenceResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return InferenceResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return resp, nil
}
