#  BlackIce Test Plan (Go + Python)

| Test ID | Layer | Scenario / Input | Expected Behavior | Metrics / Checks | Notes |
|---|---|---|---|---|---|
| UT-01 | Go Unit | Replay small PCAP (normal traffic) | Packets parsed correctly, correct flow counts | Packet count matches PCAP, no parsing errors | Start with <1k packets |
| UT-02 | Go Unit | Feature extraction test | PPS, BPS, entropy, SYN/ACK ratio computed accurately | Compare against manually computed values | Use static PCAP for reproducibility |
| UT-03 | Go Unit | IPC serialization / decoding | `FeatureWindow` marshaled + parsed correctly | Round-trip `window_id`, score decode | UDS JSON now, protobuf contract defined |
| UT-04 | Python Unit | Neural inference | Known input produces expected score | Score in [0,1], boundary cases pass | Empty + extreme windows included |
| IT-01 | IPC Integration | Go sends window → Python returns result | Correct `window_id` + `anomaly_score` | Request/response integrity; latency sample | Unix domain sockets |
| IT-02 | IPC Integration | Concurrent windows (100) | Async handling without drops | All responses returned correctly | Exercises concurrent client calls |
| E2E-01 | Replay E2E | Normal traffic | Low scores, no mitigation | Scores <0.2, allow path | Dry-run mitigation |
| E2E-02 | Replay E2E | SYN flood traffic | High scores, mitigation triggered | Scores >0.8, rate-limit/quarantine logs | Dry-run first |
| E2E-03 | Replay E2E | Mixed traffic | Scores track attack intensity | Per-window score transitions | Check false positives |
| E2E-04 | Performance | Replay 100k–300k PPS | Go fast-path keeps up; Python async | Packet loss, CPU, memory, IPC backlog | Tune pools/caps as needed |
| E2E-05 | Mitigation | Dry-run mode | Decisions logged only | Expected action logs | Enable enforcement later |
| E2E-06 | Live capture | Small Linux NIC traffic | Capture + scoring consistent | Compare with NIC counters | Sandbox first |
| ST-01 | Stress | >500k PPS replay | Measure limits | Latency, packet loss, CPU | Plan hot-loop migration if needed |
| ST-02 | Stress | 1–2 hour replay | Stable runtime | Memory trend, GC, backlog | Detect leaks/stalls |

## Execution Notes

1. Start with deterministic replay and unit tests.
2. Keep mitigation in dry-run until false positives are acceptable.
3. Record PPS/BPS, p50/p99 inference latency, memory, and CPU every run.
4. Increase load gradually and preserve artifacts in `tests/logs/`.

## Test Asset Layout

- `tests/pcap/` for replay captures (`normal_web.pcap`, `syn_flood.pcap`, `mixed_attack.pcap`)
- `tests/messages/feature_windows/` for canonical synthetic windows
- `tests/logs/` for smoke/perf artifacts
- `scripts/run_test_matrix.sh` to run automated matrix checks


## Cross-shell runners

- Bash: `./scripts/run_test_matrix.sh`
- PowerShell: `./scripts/run_test_matrix.ps1`
- CMD: `scripts\run_test_matrix.cmd`
