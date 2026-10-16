import asyncio
import logging
from concurrent.futures import ProcessPoolExecutor
from typing import Any, Dict, Optional
from uuid import UUID

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)


class VideoTaskQueue:
    """
    Менеджер очереди фоновых задач для обработки видео.
    Использует asyncio.Queue + ProcessPoolExecutor.
    """

    def __init__(self, num_workers: int = 4):
        self._queue: asyncio.Queue[Dict[str, Any]] = asyncio.Queue()
        self._jobs: Dict[str, Dict[str, Any]] = {}
        self._executor = ProcessPoolExecutor(max_workers=num_workers)
        self._workers: list[asyncio.Task] = []
        self._num_workers = num_workers

    async def start(self) -> None:
        """Запускает воркеры."""
        loop = asyncio.get_running_loop()
        for _ in range(self._num_workers):
            task = loop.create_task(self._worker_loop())
            self._workers.append(task)
        logger.info("Started %d video processor workers", self._num_workers)

    async def shutdown(self) -> None:
        """
        Корректно завершает работу:
         - дожидается опустошения очереди,
         - отменяет воркеры,
         - закрывает executor.
        """
        await self._queue.join()  # дождаться обработки всех задач
        for w in self._workers:
            w.cancel()
        await asyncio.gather(*self._workers, return_exceptions=True)
        self._executor.shutdown(wait=True)
        logger.info("VideoTaskQueue shutdown complete")

    async def add_task(self, upload_id: UUID, stage: str) -> None:
        """
        Добавляет задачу в очередь.
        """
        job = {"upload_id": upload_id, "stage": stage, "status": "queued"}
        self._jobs[upload_id] = job
        await self._queue.put(job)
        logger.info("Enqueued task %s for stage %s", upload_id, stage)

    def get_status(self, upload_id: UUID) -> Optional[Dict[str, Any]]:
        """
        Возвращает текущий статус задачи или None, если нет такого upload_id.
        """
        return self._jobs.get(upload_id)

    async def _worker_loop(self) -> None:
        """
        Вечный цикл воркера: берёт задачи, метит статус, выполняет в процессе, обновляет статус.
        """
        loop = asyncio.get_running_loop()
        while True:
            job = await self._queue.get()
            uid = job["upload_id"]
            stage = job.get("stage", "process")
            # Обновляем статус
            self._jobs[uid].update(status="processing", stage=stage)
            try:
                # здесь вызываем вашу синхронную функцию process_video
                result = await loop.run_in_executor(self._executor, process_video, job)
                self._jobs[uid].update(status="completed", result=result)
                logger.info("Task %s completed: %s", uid, result)
            except Exception as e:
                self._jobs[uid].update(status="failed", error=str(e))
                logger.exception("Task %s failed", uid)
            finally:
                self._queue.task_done()


def process_video(task_data: Dict[str, Any]) -> str:
    """
    Синхронная процедура обработки видео.
    Здесь мы просто эмулируем работу и возвращаем строку.
    """
    uid = task_data["upload_id"]
    # TODO: интегрировать реальный VideoProcessor
    return f"Processed {uid}"


task_queue = VideoTaskQueue(num_workers=4)
