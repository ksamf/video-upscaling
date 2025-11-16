import json
import asyncio
import logging
import tempfile
import signal
from confluent_kafka.experimental.aio import AIOProducer, AIOConsumer
from src.audio_processing import AudioProcessor
from src.video_processing import VideoProcessor
from config import settings


stop_event = asyncio.Event()
logger = logging.getLogger(__name__)


def handle_stop(*_):
    stop_event.set()


async def main():
    conf_consumer = {
        "bootstrap.servers": f"{settings.KAFKA_HOST}:{settings.KAFKA_PORT}",
        "group.id": "video-processor",
        "auto.offset.reset": "earliest",
    }
    conf_producer = {
        "bootstrap.servers": f"{settings.KAFKA_HOST}:{settings.KAFKA_PORT}",
    }

    consumer = AIOConsumer(conf_consumer)
    await consumer.subscribe(["video_subtitle", "video_upscale"])
    producer = AIOProducer(conf_producer)

    try:
        while not stop_event.is_set():
            msg = await consumer.poll(timeout=1.0)
            if msg is None or msg.error():
                continue

            topic = msg.topic()
            data = json.loads(msg.value().decode("utf-8"))
            if data.get("done"):
                continue

            logger.info(f"Получено сообщение из {topic}: {data}")

            if topic == "video_subtitle":
                with tempfile.TemporaryDirectory() as tmpdir:
                    audio_processor = AudioProcessor(upload_id=data["video_id"])
                    if not data["language"]:
                        orig_lang = await audio_processor.transcribe_video(tmpdir)
                        data["language"] = orig_lang
                    else:
                        await audio_processor.translate_sub(tmpdir, data["language"])
                    data["done"] = True

                    await producer.produce(
                        topic="video_subtitle",
                        value=json.dumps(data).encode("utf-8"),
                    )
                    await producer.flush()

            elif topic == "video_upscale":
                with tempfile.TemporaryDirectory() as tmpdir:
                    video_processor = VideoProcessor(
                        upload_id=data["video_id"],
                        realistic=data["realistic"],
                    )
                    await video_processor.upscale_video(
                        data["video_id"], data["quality"], tmpdir
                    )
                    data["done"] = True

                    await producer.produce(
                        topic="video_upscale",
                        value=json.dumps(data).encode("utf-8"),
                    )
                    await producer.flush()
    finally:
        await consumer.close()
        await producer.close()


if __name__ == "__main__":
    signal.signal(signal.SIGINT, handle_stop)
    signal.signal(signal.SIGTERM, handle_stop)
    asyncio.run(main())
