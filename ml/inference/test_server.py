import unittest

from server import score


class ScoreTests(unittest.TestCase):
    def test_score_bounds(self):
        self.assertGreaterEqual(score({}), 0.0)
        self.assertLessEqual(score({}), 1.0)

    def test_empty_window_low_score(self):
        result = score(
            {
                "pps": 0,
                "syn_ack_ratio": 0,
                "flow_churn": 0,
                "dst_port_entropy": 2.0,
            }
        )
        self.assertLess(result, 0.1)

    def test_extreme_window_high_score(self):
        result = score(
            {
                "pps": 600000,
                "syn_ack_ratio": 30,
                "flow_churn": 1.0,
                "dst_port_entropy": 0.1,
            }
        )
        self.assertGreaterEqual(result, 0.95)


if __name__ == "__main__":
    unittest.main()
