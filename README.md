# BlackIce (Go + Python MVP)

BlackIce uses a split runtime:

- **Go core runtime** for capture, fast-path feature extraction, and mitigation.
- **Python sidecar** for anomaly scoring/inference.
- **Unix domain socket IPC** today, with a protobuf contract for migration to binary protobuf/gRPC.

## Run

Terminal 1:

```bash
python3 ml/inference/server.py --socket /tmp/blackice.sock
```

Terminal 2:

```bash
go run ./cmd/blackice --socket /tmp/blackice.sock --pps 120000 --window 1s
```

## Architecture

```text
NIC/PCAP -> Go Capture -> Fast Path -> Feature Exporter -> Python Inference -> Go Mitigation
```

## Package layout

```text
blackice/
├── cmd/blackice
├── core/
│   ├── capture
│   ├── fastpath
│   ├── features
│   ├── flows
│   ├── ipc
│   ├── mitigation
│   └── runtime
├── ml/
│   ├── inference
│   ├── model
│   └── training
├── proto
├── replay
├── configs
└── benchmarks
```

## Testing

Run the full local matrix (Go tests, Python tests, and sidecar/runtime smoke):

```bash
./scripts/run_test_matrix.sh
```

```powershell
./scripts/run_test_matrix.ps1
```

```cmd
scripts\run_test_matrix.cmd
```

Test assets live under `tests/`:

- `tests/pcap/` replay fixtures
- `tests/messages/feature_windows/` canonical feature windows
- `tests/logs/` generated test artifacts

