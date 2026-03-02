package capture

import (
	"context"
	"math/rand"
	"time"
)

// Packet is the minimum structure needed by the fast path.
type Packet struct {
	Timestamp time.Time
	SrcIP     string
	DstPort   uint16
	Bytes     int
	SYN       bool
	ACK       bool
}

// Source is a packet producer that can feed the runtime loop.
type Source interface {
	Run(ctx context.Context, out chan<- Packet)
}

// ReplaySource simulates traffic when PCAP replay is not wired yet.
type ReplaySource struct {
	RatePerSecond int
}

func (r ReplaySource) Run(ctx context.Context, out chan<- Packet) {
	rate := r.RatePerSecond
	if rate <= 0 {
		rate = 1000
	}
	interval := time.Second / time.Duration(rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			out <- Packet{
				Timestamp: now,
				SrcIP:     randomIP(),
				DstPort:   randomPort(),
				Bytes:     60 + rand.Intn(1400),
				SYN:       rand.Intn(10) < 3,
				ACK:       rand.Intn(10) < 8,
			}
		}
	}
}

func randomIP() string {
	return "10.0." + itoa(1+rand.Intn(3)) + "." + itoa(1+rand.Intn(254))
}

func randomPort() uint16 {
	ports := []uint16{53, 80, 123, 443, 8080, 8443}
	return ports[rand.Intn(len(ports))]
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [3]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
