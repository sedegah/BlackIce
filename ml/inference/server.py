#!/usr/bin/env python3
import argparse
import json
import os
import socket


def score(features: dict) -> float:
    pps = float(features.get("pps", 0.0))
    syn_ack = float(features.get("syn_ack_ratio", 0.0))
    churn = float(features.get("flow_churn", 0.0))
    dst_entropy = float(features.get("dst_port_entropy", 0.0))

    score = 0.0
    score += min(0.4, pps / 300000.0 * 0.4)
    score += min(0.3, syn_ack / 10.0 * 0.3)
    score += min(0.2, churn / 1.0 * 0.2)
    score += max(0.0, (1.5 - dst_entropy) / 1.5 * 0.1)
    return max(0.0, min(1.0, score))


def serve(socket_path: str) -> None:
    if os.path.exists(socket_path):
        os.remove(socket_path)

    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as srv:
        srv.bind(socket_path)
        srv.listen(128)
        print(f"blackice python inference listening on {socket_path}")

        while True:
            conn, _ = srv.accept()
            with conn:
                payload = conn.recv(8192)
                if not payload:
                    continue
                req = json.loads(payload.decode("utf-8"))
                resp = {
                    "window_id": req["window_id"],
                    "anomaly_score": round(score(req), 4),
                }
                conn.sendall((json.dumps(resp) + "\n").encode("utf-8"))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="BlackIce Python inference sidecar")
    parser.add_argument("--socket", default="/tmp/blackice.sock")
    args = parser.parse_args()
    serve(args.socket)
