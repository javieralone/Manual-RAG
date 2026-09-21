import io
import json
import logging
import unittest

from app.observability import JsonFormatter, trace_id_context


class ObservabilityTests(unittest.TestCase):
    def test_json_formatter_includes_trace_id_without_sensitive_payload(self):
        stream = io.StringIO()
        handler = logging.StreamHandler(stream)
        handler.setFormatter(JsonFormatter())
        logger = logging.getLogger("observability-test")
        logger.handlers = [handler]
        logger.setLevel(logging.INFO)
        logger.propagate = False
        token = trace_id_context.set("trace-id-test")
        try:
            logger.info("request_completed")
        finally:
            trace_id_context.reset(token)
        payload = json.loads(stream.getvalue())
        self.assertEqual(payload["trace_id"], "trace-id-test")
        self.assertEqual(payload["message"], "request_completed")
        self.assertNotIn("Authorization", payload)


if __name__ == "__main__":
    unittest.main()
