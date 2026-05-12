import redis
import json
import logging
import threading

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def start_subscriber():
    def _subscribe():
        client = redis.Redis(host="localhost", port=6379, db=0)
        pubsub = client.pubsub()
        pubsub.subscribe("driver:updates")

        logger.info("Python subscriber listening on driver:updates")

        for message in pubsub.listen():
            if message["type"] == "message":
                try:
                    data = json.loads(message["data"])
                    logger.info(
                        f"Received: driver={data['driver_id']} "
                        f"lat={data['latitude']:.4f} "
                        f"lng={data['longitude']:.4f}"
                    )
                except json.JSONDecodeError as e:
                    logger.error(f"Failed to parse message: {e}")

    thread = threading.Thread(target=_subscribe, daemon=True)
    thread.start()
    logger.info("Subscriber thread started")