package fastpath

import (
	"math"

	"blackice/core/capture"
)

type WindowFeatures struct {
	WindowID        uint64  `json:"window_id"`
	PPS             int     `json:"pps"`
	BPS             int     `json:"bps"`
	SrcIPEntropy    float64 `json:"src_ip_entropy"`
	DstPortEntropy  float64 `json:"dst_port_entropy"`
	SynAckRatio     float64 `json:"syn_ack_ratio"`
	FlowChurn       float64 `json:"flow_churn"`
	SuspiciousScore float64 `json:"suspicious_score"`
}

type Aggregator struct {
	windowPackets int
	windowBytes   int
	synCount      int
	ackCount      int
	srcIPCounts   map[string]int
	dstPortCounts map[uint16]int
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		srcIPCounts:   make(map[string]int, 512),
		dstPortCounts: make(map[uint16]int, 64),
	}
}

func (a *Aggregator) Add(pkt capture.Packet) {
	a.windowPackets++
	a.windowBytes += pkt.Bytes
	if pkt.SYN {
		a.synCount++
	}
	if pkt.ACK {
		a.ackCount++
	}
	a.srcIPCounts[pkt.SrcIP]++
	a.dstPortCounts[pkt.DstPort]++
}

func (a *Aggregator) Close(windowID uint64) WindowFeatures {
	pps := a.windowPackets
	bps := a.windowBytes * 8
	f := WindowFeatures{
		WindowID:       windowID,
		PPS:            pps,
		BPS:            bps,
		SrcIPEntropy:   entropyIntMap(a.srcIPCounts),
		DstPortEntropy: entropyPortMap(a.dstPortCounts),
		SynAckRatio:    ratio(a.synCount, a.ackCount),
		FlowChurn:      float64(len(a.srcIPCounts)) / float64(max(1, pps)),
	}
	f.SuspiciousScore = heuristicScore(f)
	a.reset()
	return f
}

func (a *Aggregator) reset() {
	a.windowPackets, a.windowBytes = 0, 0
	a.synCount, a.ackCount = 0, 0
	for k := range a.srcIPCounts {
		delete(a.srcIPCounts, k)
	}
	for k := range a.dstPortCounts {
		delete(a.dstPortCounts, k)
	}
}

func entropyIntMap(m map[string]int) float64 {
	total := 0
	for _, c := range m {
		total += c
	}
	if total == 0 {
		return 0
	}
	return entropy(total, func(yield func(float64) bool) {
		for _, c := range m {
			if !yield(float64(c) / float64(total)) {
				return
			}
		}
	})
}

func entropyPortMap(m map[uint16]int) float64 {
	total := 0
	for _, c := range m {
		total += c
	}
	if total == 0 {
		return 0
	}
	return entropy(total, func(yield func(float64) bool) {
		for _, c := range m {
			if !yield(float64(c) / float64(total)) {
				return
			}
		}
	})
}

func entropy(_ int, iter func(func(float64) bool)) float64 {
	h := 0.0
	iter(func(p float64) bool {
		h += -p * math.Log2(p)
		return true
	})
	return h
}

func ratio(a, b int) float64 {
	if b == 0 {
		return float64(a)
	}
	return float64(a) / float64(b)
}

func heuristicScore(f WindowFeatures) float64 {
	score := 0.0
	if f.PPS > 100000 {
		score += 0.35
	}
	if f.SynAckRatio > 5 {
		score += 0.25
	}
	if f.DstPortEntropy < 1 {
		score += 0.2
	}
	if f.FlowChurn > 0.6 {
		score += 0.2
	}
	if score > 1 {
		return 1
	}
	return score
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
