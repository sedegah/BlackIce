# BlackIce MVP Test Plan

| Level | Scenario | Input | Expected |
|---|---|---|---|
| Unit (Go) | Aggregator stats | Synthetic packets | Correct PPS/BPS, entropy, SYN/ACK, churn |
| Unit (Go) | IPC response decoding | Local UDS mock | window_id echo + anomaly parsed |
| Unit (Py) | Score bounds | empty/extreme feature windows | score always 0..1 |
| Integration | Runtime->IPC | Fixed high-risk packet stream | at least one inference call |
| E2E smoke | Python sidecar + Go runtime | simulated replay | mitigation decisions logged |

## Suggested PCAP Scenarios (next step)

1. Baseline web workload (`normal_web.pcap`) – expected anomaly `< 0.3`
2. SYN flood (`syn_flood.pcap`) – expected anomaly `> 0.8`, quarantine/rate-limit
3. Mixed traffic (`mixed_attack.pcap`) – anomaly spikes during attack windows

## Performance Checklist

- PPS/BPS per window
- Go heap growth and GC pauses
- Python inference latency (p50/p99)
- UDS request backlog
