package fastpath

import (
	"math"
	"testing"
	"time"

	"blackice/core/capture"
)

func TestAggregatorComputesExpectedWindowFeatures(t *testing.T) {
	agg := NewAggregator()
	now := time.Unix(1700000000, 0)
	pkts := []capture.Packet{
		{Timestamp: now, SrcIP: "10.0.1.1", DstPort: 80, Bytes: 100, SYN: true, ACK: false},
		{Timestamp: now, SrcIP: "10.0.1.1", DstPort: 80, Bytes: 100, SYN: true, ACK: true},
		{Timestamp: now, SrcIP: "10.0.1.2", DstPort: 443, Bytes: 50, SYN: false, ACK: true},
		{Timestamp: now, SrcIP: "10.0.1.3", DstPort: 443, Bytes: 50, SYN: false, ACK: true},
	}
	for _, p := range pkts {
		agg.Add(p)
	}

	f := agg.Close(9)
	if f.WindowID != 9 {
		t.Fatalf("window id mismatch: got %d", f.WindowID)
	}
	if f.PPS != 4 {
		t.Fatalf("pps mismatch: got %d", f.PPS)
	}
	if f.BPS != 2400 {
		t.Fatalf("bps mismatch: got %d", f.BPS)
	}
	if diff := math.Abs(f.SynAckRatio - (2.0 / 3.0)); diff > 1e-6 {
		t.Fatalf("syn/ack mismatch: got %.6f", f.SynAckRatio)
	}
	if diff := math.Abs(f.FlowChurn - 0.75); diff > 1e-6 {
		t.Fatalf("flow churn mismatch: got %.6f", f.FlowChurn)
	}
	if !(f.SrcIPEntropy > 1.4 && f.SrcIPEntropy < 1.6) {
		t.Fatalf("src entropy unexpected: %.6f", f.SrcIPEntropy)
	}
	if !(f.DstPortEntropy > 0.9 && f.DstPortEntropy < 1.1) {
		t.Fatalf("dst entropy unexpected: %.6f", f.DstPortEntropy)
	}
}

func TestAggregatorResetAfterClose(t *testing.T) {
	agg := NewAggregator()
	agg.Add(capture.Packet{SrcIP: "10.0.1.1", DstPort: 80, Bytes: 64, SYN: true})
	_ = agg.Close(1)
	f2 := agg.Close(2)
	if f2.PPS != 0 || f2.BPS != 0 {
		t.Fatalf("expected reset window, got pps=%d bps=%d", f2.PPS, f2.BPS)
	}
}
