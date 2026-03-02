package runtime

import (
	"context"
	"log"
	"time"

	"blackice/core/capture"
	"blackice/core/fastpath"
	"blackice/core/features"
	"blackice/core/ipc"
	"blackice/core/mitigation"
)

type Config struct {
	Window           time.Duration
	PacketRate       int
	PythonSocket     string
	InferenceTimeout time.Duration
	Source           capture.Source
}

type Service struct {
	cfg       Config
	agg       *fastpath.Aggregator
	ipc       *ipc.Client
	mitigator mitigation.Engine
}

func NewService(cfg Config) *Service {
	return &Service{
		cfg: cfg,
		agg: fastpath.NewAggregator(),
		ipc: ipc.NewClient(cfg.PythonSocket, cfg.InferenceTimeout),
		mitigator: mitigation.Engine{
			QuarantineThreshold: 0.8,
		},
	}
}

func (s *Service) Run(ctx context.Context) error {
	packets := make(chan capture.Packet, 4096)
	source := s.cfg.Source
	if source == nil {
		source = capture.ReplaySource{RatePerSecond: s.cfg.PacketRate}
	}
	go source.Run(ctx, packets)

	tick := time.NewTicker(s.cfg.Window)
	defer tick.Stop()
	windowID := uint64(1)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case pkt := <-packets:
			s.agg.Add(pkt)
		case <-tick.C:
			f := s.agg.Close(windowID)
			windowID++
			if !features.ShouldEscalate(f) {
				s.mitigator.Decide(f, ipc.InferenceResponse{WindowID: f.WindowID, AnomalyScore: f.SuspiciousScore})
				continue
			}
			resp, err := s.ipc.Infer(ctx, f)
			if err != nil {
				log.Printf("python inference unavailable, fallback score=%.2f err=%v", f.SuspiciousScore, err)
				resp = ipc.InferenceResponse{WindowID: f.WindowID, AnomalyScore: f.SuspiciousScore}
			}
			s.mitigator.Decide(f, resp)
		}
	}
}
