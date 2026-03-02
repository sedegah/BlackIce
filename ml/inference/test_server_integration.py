import json
import os
import socket
import tempfile
import threading
import time
import unittest

from server import serve


class ServerIntegrationTests(unittest.TestCase):
    def test_unix_socket_round_trip(self):
        with tempfile.TemporaryDirectory() as d:
            sock_path = os.path.join(d, "blackice.sock")
            thread = threading.Thread(target=serve, args=(sock_path,), daemon=True)
            thread.start()

            deadline = time.time() + 2
            while not os.path.exists(sock_path):
                if time.time() > deadline:
                    self.fail("socket did not appear")
                time.sleep(0.02)

            req = {
                "window_id": 22,
                "pps": 250000,
                "syn_ack_ratio": 10,
                "flow_churn": 0.9,
                "dst_port_entropy": 0.2,
            }
            with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as c:
                c.connect(sock_path)
                c.sendall((json.dumps(req) + "\n").encode("utf-8"))
                raw = c.recv(4096)

            resp = json.loads(raw.decode("utf-8").strip())
            self.assertEqual(resp["window_id"], 22)
            self.assertGreaterEqual(resp["anomaly_score"], 0.0)
            self.assertLessEqual(resp["anomaly_score"], 1.0)


if __name__ == "__main__":
    unittest.main()
