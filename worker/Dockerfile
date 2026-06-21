FROM python:3.12-slim

# Unbuffered output so structured logs flush immediately to docker/journald.
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY main.py .

# Exec form => python is PID 1 and receives SIGTERM/SIGINT directly, so the
# worker can finish its in-flight job and exit cleanly on `docker stop`.
# The compose `worker` service sets `restart: unless-stopped`.
# Tunables come from EDU_* env (see docker-compose.yml).
STOPSIGNAL SIGTERM
ENTRYPOINT ["python", "main.py"]
