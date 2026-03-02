package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os/signal"
	"syscall"
	"time"

	"blackice/core/runtime"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	socket := flag.String("socket", "/tmp/blackice.sock", "python inference unix socket")
	window := flag.Duration("window", 1*time.Second, "feature aggregation window")
	rate := flag.Int("pps", 10000, "simulated packet rate")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	svc := runtime.NewService(runtime.Config{
		Window:           *window,
		PacketRate:       *rate,
		PythonSocket:     *socket,
		InferenceTimeout: 250 * time.Millisecond,
	})

	if err := svc.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
