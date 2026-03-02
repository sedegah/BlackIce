package mitigation

import (
	"fmt"
	"log"

	"blackice/core/fastpath"
	"blackice/core/ipc"
)

type Engine struct {
	QuarantineThreshold float64
}

func (e Engine) Decide(features fastpath.WindowFeatures, inference ipc.InferenceResponse) {
	score := inference.AnomalyScore
	if score == 0 {
		score = features.SuspiciousScore
	}
	if score >= e.QuarantineThreshold {
		log.Printf("mitigation=quarantine window=%d score=%.2f", features.WindowID, score)
		return
	}
	if score >= 0.45 {
		log.Printf("mitigation=rate-limit window=%d score=%.2f", features.WindowID, score)
		return
	}
	fmt.Printf("mitigation=allow window=%d score=%.2f\n", features.WindowID, score)
}
