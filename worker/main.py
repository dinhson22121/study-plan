from __future__ import annotations

import dataclasses
import json
import logging
import os
import re
import signal
import time
import uuid

import boto3
import fitz
import psycopg
from botocore.config import Config as BotoConfig
from botocore.exceptions import BotoCoreError, ClientError

try:
    import sentry_sdk
except ImportError:  # sentry is optional; worker runs fine without it
    sentry_sdk = None


# --- structured logging ------------------------------------------------------

class KeyValueFormatter(logging.Formatter):
    """Render log records as `key=value` pairs for easy ingestion.

    Standard fields (ts, level, logger, msg) come first; any `extra=` keys
    passed on the record are appended so per-job context (job id, asset id,
    duration, outcome) shows up in every line.
    """

    _RESERVED = set(
        logging.makeLogRecord({}).__dict__.keys()
    ) | {"message", "asctime", "taskName"}

    def format(self, record: logging.LogRecord) -> str:
        base = {
            "ts": self.formatTime(record, "%Y-%m-%dT%H:%M:%S%z"),
            "level": record.levelname,
            "logger": record.name,
            "msg": record.getMessage(),
        }
        extras = {
            k: v
            for k, v in record.__dict__.items()
            if k not in self._RESERVED and not k.startswith("_")
        }
        parts = [f"{k}={_fmt_val(v)}" for k, v in {**base, **extras}.items()]
        line = " ".join(parts)
        if record.exc_info:
            line += " exc=" + json.dumps(self.formatException(record.exc_info))
        return line


def _fmt_val(value) -> str:
    s = str(value)
    if s == "" or any(c in s for c in " \t\n\"="):
        return json.dumps(s)
    return s


def _configure_logging() -> logging.Logger:
    handler = logging.StreamHandler()
    handler.setFormatter(KeyValueFormatter())
    root = logging.getLogger()
    root.handlers = [handler]
    root.setLevel(os.getenv("EDU_PARSER_LOG_LEVEL", "INFO").upper())
    return logging.getLogger("pdf_parser")


log = _configure_logging()


# --- configuration -----------------------------------------------------------

MAX_PDF_BYTES = 20 * 1024 * 1024  # mirrors the 20MB admin upload limit


@dataclasses.dataclass(frozen=True)
class Config:
    dsn: str
    s3_endpoint: str
    s3_region: str
    s3_access_key: str
    s3_secret_key: str
    s3_bucket: str
    s3_path_style: bool
    poll_interval: float
    parser_version: str
    stuck_timeout_seconds: float
    max_attempts: int
    max_retries: int
    retry_backoff_base: float
    sentry_dsn: str

    @staticmethod
    def from_env() -> "Config":
        return Config(
            dsn=os.environ["EDU_POSTGRES_URL"],
            s3_endpoint=os.getenv("EDU_S3_ENDPOINT", ""),
            s3_region=os.getenv("EDU_S3_REGION", "us-east-1"),
            s3_access_key=os.getenv("EDU_S3_ACCESS_KEY", ""),
            s3_secret_key=os.getenv("EDU_S3_SECRET_KEY", ""),
            s3_bucket=os.environ["EDU_S3_BUCKET"],
            s3_path_style=os.getenv("EDU_S3_USE_PATH_STYLE", "true").lower() == "true",
            poll_interval=float(os.getenv("EDU_PARSER_POLL_INTERVAL", "5")),
            parser_version=os.getenv("EDU_PARSER_VERSION", "pdf-parser-mvp-1"),
            stuck_timeout_seconds=float(os.getenv("EDU_PARSER_STUCK_TIMEOUT", "600")),
            max_attempts=int(os.getenv("EDU_PARSER_MAX_ATTEMPTS", "3")),
            max_retries=int(os.getenv("EDU_PARSER_MAX_RETRIES", "3")),
            retry_backoff_base=float(os.getenv("EDU_PARSER_RETRY_BACKOFF", "1.5")),
            sentry_dsn=os.getenv("EDU_SENTRY_DSN", ""),
        )


def s3_client(cfg: Config):
    kwargs = {
        "region_name": cfg.s3_region,
        "aws_access_key_id": cfg.s3_access_key,
        "aws_secret_access_key": cfg.s3_secret_key,
        "config": BotoConfig(
            s3={"addressing_style": "path" if cfg.s3_path_style else "auto"},
            retries={"max_attempts": cfg.max_retries, "mode": "standard"},
        ),
    }
    if cfg.s3_endpoint:
        kwargs["endpoint_url"] = cfg.s3_endpoint
    return boto3.client("s3", **kwargs)


WORKER_ID = f"pdf-parser-{uuid.uuid4().hex[:8]}"


# --- error taxonomy ----------------------------------------------------------

class TransientError(Exception):
    """Recoverable infrastructure error (S3/DB hiccup) — retry with backoff."""


class PermanentError(Exception):
    """Unrecoverable input error (corrupt/unsupported PDF) — fail the job."""


# psycopg connection/operational errors we treat as transient
_TRANSIENT_DB_ERRORS = (
    psycopg.OperationalError,
    psycopg.InterfaceError,
)


# --- graceful shutdown -------------------------------------------------------

class ShutdownFlag:
    """Set by SIGTERM/SIGINT. The loop checks it between jobs so the in-flight
    job always finishes before the process exits cleanly."""

    def __init__(self) -> None:
        self._requested = False

    def request(self, signum, _frame) -> None:
        self._requested = True
        log.info("shutdown requested", extra={"signal": signal.Signals(signum).name})

    @property
    def requested(self) -> bool:
        return self._requested


# --- DB job lifecycle --------------------------------------------------------

def reset_stuck_jobs(conn, cfg: Config) -> int:
    """Recover jobs stuck in PROCESSING past the timeout.

    The schema has no PENDING status, so recoverable jobs go back to QUEUED.
    Jobs that have already burned through max_attempts are marked FAILED so
    they cannot loop forever. Uses updated_at as the heartbeat.
    """
    threshold = f"{int(cfg.stuck_timeout_seconds)} seconds"
    with conn.cursor() as cur:
        cur.execute(
            """UPDATE parse_job
               SET status='FAILED', finished_at=now(), updated_at=now(),
                   error_message=%s
               WHERE status='PROCESSING'
                 AND updated_at < now() - %s::interval
                 AND attempt_count >= %s""",
            (
                "exceeded max attempts after stalling in PROCESSING",
                threshold,
                cfg.max_attempts,
            ),
        )
        failed = cur.rowcount
        cur.execute(
            """UPDATE parse_job
               SET status='QUEUED', claimed_by=NULL, claimed_at=NULL,
                   started_at=NULL, updated_at=now()
               WHERE status='PROCESSING'
                 AND updated_at < now() - %s::interval
                 AND attempt_count < %s""",
            (threshold, cfg.max_attempts),
        )
        requeued = cur.rowcount
    conn.commit()
    if failed or requeued:
        log.warning(
            "recovered stuck jobs",
            extra={"requeued": requeued, "failed": failed, "timeout_s": cfg.stuck_timeout_seconds},
        )
    return requeued + failed


def claim_job(conn, parser_version: str):
    sql = """
        WITH j AS (
            SELECT id FROM parse_job
            WHERE status = 'QUEUED'
            ORDER BY created_at
            FOR UPDATE SKIP LOCKED
            LIMIT 1
        )
        UPDATE parse_job p
        SET status='PROCESSING', claimed_by=%s, claimed_at=now(),
            started_at=now(), attempt_count=attempt_count+1,
            parser_version=%s, updated_at=now()
        FROM j WHERE p.id = j.id
        RETURNING p.id, p.asset_id, p.attempt_count
    """
    with conn.cursor() as cur:
        cur.execute(sql, (WORKER_ID, parser_version))
        row = cur.fetchone()
    conn.commit()
    return row


def fetch_asset(conn, asset_id):
    with conn.cursor() as cur:
        cur.execute(
            """SELECT object_key, bucket_name, content_type, file_size, original_filename
               FROM uploaded_asset WHERE id = %s""",
            (asset_id,),
        )
        return cur.fetchone()


def finish_job(conn, job_id, status, raw_text="", error="") -> None:
    with conn.cursor() as cur:
        cur.execute(
            """UPDATE parse_job
               SET status=%s, finished_at=now(), updated_at=now(),
                   raw_text=%s, error_message=%s
               WHERE id=%s""",
            (status, raw_text[:200000], error[:2000], job_id),
        )
    conn.commit()


def requeue_job(conn, job_id, error="") -> None:
    """Return a job to QUEUED after a transient failure so it is retried later
    by this or another worker. attempt_count was already incremented at claim."""
    with conn.cursor() as cur:
        cur.execute(
            """UPDATE parse_job
               SET status='QUEUED', claimed_by=NULL, claimed_at=NULL,
                   started_at=NULL, updated_at=now(), error_message=%s
               WHERE id=%s""",
            (error[:2000], job_id),
        )
    conn.commit()


# --- parsing -----------------------------------------------------------------

QUESTION_RE = re.compile(r"(?:^|\n)\s*(?:Câu|Bài|Question)\s*(\d+)\s*[:.)]?\s*", re.IGNORECASE)
OPTION_RE = re.compile(r"(?:^|\n)\s*([A-D])\s*[.)]\s*(.+)")
ANSWER_RE = re.compile(r"(?:Đáp\s*án|Answer|ĐA)\s*[:.]?\s*([A-D])", re.IGNORECASE)

PDF_MAGIC = b"%PDF-"
MIN_MEANINGFUL_TEXT = 20  # chars of non-whitespace text below which we treat as image-only


@dataclasses.dataclass
class ParsedQuestion:
    number: int
    stem: str
    options: list
    confidence: float


def validate_pdf(pdf_bytes: bytes, content_type: str, filename: str) -> None:
    """Enforce the supported-format policy before extraction.

    Raises PermanentError for anything we cannot meaningfully parse so the job
    fails with an admin-readable message instead of crashing or looping.
    """
    if not pdf_bytes:
        raise PermanentError("uploaded file is empty (0 bytes)")
    if len(pdf_bytes) > MAX_PDF_BYTES:
        raise PermanentError(
            f"file is {len(pdf_bytes)} bytes, exceeds the {MAX_PDF_BYTES}-byte limit"
        )
    if not pdf_bytes.startswith(PDF_MAGIC):
        ct = content_type or "unknown"
        raise PermanentError(
            f"file is not a PDF (content_type={ct!r}, filename={filename!r}); "
            "only text-based PDFs are supported"
        )


def extract_text(pdf_bytes: bytes) -> str:
    """Open the PDF and concatenate page text.

    A failure to open is a permanent (corrupt/unsupported) error; an empty or
    near-empty result means an image-only / scanned PDF, which we do not OCR.
    """
    try:
        doc = fitz.open(stream=pdf_bytes, filetype="pdf")
    except Exception as exc:  # PyMuPDF raises various low-level errors here
        raise PermanentError(f"could not open PDF (corrupt or unsupported): {exc}") from exc
    try:
        if doc.page_count == 0:
            raise PermanentError("PDF has no pages")
        text = "\n".join(page.get_text() for page in doc)
    finally:
        doc.close()

    if len(text.strip()) < MIN_MEANINGFUL_TEXT:
        raise PermanentError(
            "PDF contains little or no extractable text (likely scanned/image-only); "
            "OCR is not supported"
        )
    return text


def parse_questions(text: str) -> list:
    parts = QUESTION_RE.split(text)
    out = []
    for i in range(1, len(parts), 2):
        try:
            number = int(parts[i])
        except (ValueError, IndexError):
            continue
        body = parts[i + 1] if i + 1 < len(parts) else ""
        q = _parse_block(number, body)
        if q is not None:
            out.append(q)
    return out


def _parse_block(number: int, body: str):
    answer_label = None
    m = ANSWER_RE.search(body)
    if m:
        answer_label = m.group(1).upper()

    options = []
    first_opt_pos = None
    for om in OPTION_RE.finditer(body):
        if first_opt_pos is None:
            first_opt_pos = om.start()
        label = om.group(1).upper()
        opt_text = om.group(2).strip()
        opt_text = ANSWER_RE.sub("", opt_text).strip()
        options.append([label, opt_text, label == answer_label])

    stem = (body[:first_opt_pos] if first_opt_pos is not None else body).strip()
    if not stem:
        return None

    confidence = 0.4
    if len(options) >= 2:
        confidence = 0.7
    if len(options) == 4:
        confidence = 0.85
    if answer_label and any(o[2] for o in options):
        confidence = min(1.0, confidence + 0.15)

    return ParsedQuestion(number=number, stem=stem, options=options, confidence=round(confidence, 3))


def save_drafts(conn, job_id, asset_id, questions: list) -> None:
    with conn.cursor() as cur:
        cur.execute("DELETE FROM question_draft WHERE parse_job_id = %s", (job_id,))
        for q in questions:
            draft_id = str(uuid.uuid4())
            cur.execute(
                """INSERT INTO question_draft
                   (id, asset_id, parse_job_id, question_number, question_type, stem,
                    explanation_raw, answer_key_raw, parse_confidence, status, created_at, updated_at)
                   VALUES (%s,%s,%s,%s,'MCQ',%s,'', %s,%s,'DRAFT',now(),now())""",
                (draft_id, asset_id, job_id, q.number, q.stem,
                 _answer_key(q), q.confidence),
            )
            for idx, (label, text, correct) in enumerate(q.options):
                cur.execute(
                    """INSERT INTO question_draft_option
                       (id, question_draft_id, option_label, option_text, is_correct_inferred, order_index)
                       VALUES (%s,%s,%s,%s,%s,%s)""",
                    (str(uuid.uuid4()), draft_id, label, text, correct, idx),
                )
    conn.commit()


def _answer_key(q: ParsedQuestion) -> str:
    return ",".join(o[0] for o in q.options if o[2])


def needs_review(questions: list) -> bool:
    if not questions:
        return True
    for q in questions:
        if len(q.options) < 2 or not any(o[2] for o in q.options):
            return True
    return False


# --- job processing ----------------------------------------------------------

def _download_pdf(s3, cfg: Config, object_key: str, bucket: str) -> bytes:
    """Download from S3, translating infra failures into TransientError and
    bounding memory by refusing anything over the upload limit."""
    target_bucket = bucket or cfg.s3_bucket
    try:
        head = s3.head_object(Bucket=target_bucket, Key=object_key)
        size = head.get("ContentLength", 0)
        if size > MAX_PDF_BYTES:
            raise PermanentError(
                f"object is {size} bytes, exceeds the {MAX_PDF_BYTES}-byte limit"
            )
        obj = s3.get_object(Bucket=target_bucket, Key=object_key)
        return obj["Body"].read(MAX_PDF_BYTES + 1)
    except ClientError as exc:
        code = exc.response.get("Error", {}).get("Code", "")
        if code in ("NoSuchKey", "404", "NoSuchBucket"):
            raise PermanentError(f"object not found in storage: {object_key}") from exc
        raise TransientError(f"S3 error fetching object: {exc}") from exc
    except BotoCoreError as exc:
        raise TransientError(f"S3 connection error: {exc}") from exc


def _process_claimed(conn, s3, cfg: Config, job_id, asset_id, attempt_count: int) -> None:
    """Run the actual parse for an already-claimed job. Raises TransientError
    (caller may requeue/retry) or PermanentError (caller marks FAILED)."""
    asset = fetch_asset(conn, asset_id)
    if not asset:
        raise PermanentError("asset row not found")
    object_key, bucket, content_type, _file_size, filename = asset

    pdf_bytes = _download_pdf(s3, cfg, object_key, bucket)
    validate_pdf(pdf_bytes, content_type, filename or "")

    text = extract_text(pdf_bytes)
    questions = parse_questions(text)
    save_drafts(conn, job_id, asset_id, questions)

    status = "REVIEW_REQUIRED" if needs_review(questions) else "PARSED"
    finish_job(conn, job_id, status, raw_text=text)
    log.info(
        "job complete",
        extra={"job_id": job_id, "asset_id": asset_id, "status": status,
               "drafts": len(questions), "attempt": attempt_count},
    )


def process_one(conn, s3, cfg: Config) -> bool:
    """Claim and process a single job. Returns True if a job was handled (so the
    drain loop keeps going), False when the queue is empty."""
    claimed = claim_job(conn, cfg.parser_version)
    if not claimed:
        return False
    job_id, asset_id, attempt_count = claimed
    started = time.monotonic()
    log.info("job claimed", extra={"job_id": job_id, "asset_id": asset_id, "attempt": attempt_count})

    try:
        _process_claimed(conn, s3, cfg, job_id, asset_id, attempt_count)
    except PermanentError as exc:
        _safe_rollback(conn)
        finish_job(conn, job_id, "FAILED", error=str(exc))
        log.warning(
            "job failed (permanent)",
            extra={"job_id": job_id, "asset_id": asset_id, "error": str(exc),
                   "duration_s": round(time.monotonic() - started, 3)},
        )
    except TransientError as exc:
        _safe_rollback(conn)
        _capture(exc)
        if attempt_count >= cfg.max_attempts:
            finish_job(conn, job_id, "FAILED",
                       error=f"transient error, gave up after {attempt_count} attempts: {exc}")
            log.error(
                "job failed (transient, exhausted)",
                extra={"job_id": job_id, "asset_id": asset_id, "error": str(exc),
                       "attempt": attempt_count},
            )
        else:
            requeue_job(conn, job_id, error=str(exc))
            log.warning(
                "job requeued (transient)",
                extra={"job_id": job_id, "asset_id": asset_id, "error": str(exc),
                       "attempt": attempt_count},
            )
    except _TRANSIENT_DB_ERRORS:
        # DB dropped mid-job: let it bubble so the outer loop reconnects. The
        # job stays PROCESSING and will be recovered by reset_stuck_jobs.
        raise
    except Exception as exc:
        _safe_rollback(conn)
        _capture(exc)
        finish_job(conn, job_id, "FAILED", error=f"unexpected error: {exc}")
        log.exception("job failed (unexpected)", extra={"job_id": job_id, "asset_id": asset_id})
    return True


def _safe_rollback(conn) -> None:
    try:
        conn.rollback()
    except Exception:
        log.warning("rollback failed", extra={"worker_id": WORKER_ID})


def _safe_close(conn) -> None:
    try:
        if conn is not None and not conn.closed:
            conn.close()
    except Exception:
        pass


def _capture(exc: Exception) -> None:
    if sentry_sdk is not None:
        sentry_sdk.capture_exception(exc)


# --- main loop ---------------------------------------------------------------

def _init_sentry(cfg: Config) -> None:
    if not cfg.sentry_dsn:
        return
    if sentry_sdk is None:
        log.warning("EDU_SENTRY_DSN set but sentry-sdk not installed; skipping")
        return
    sentry_sdk.init(dsn=cfg.sentry_dsn, traces_sample_rate=0.0)
    log.info("sentry enabled")


def _connect_with_backoff(cfg: Config, shutdown: ShutdownFlag):
    """Establish a DB connection, retrying transient failures with bounded
    exponential backoff. Returns None if shutdown is requested first."""
    attempt = 0
    while not shutdown.requested:
        try:
            return psycopg.connect(cfg.dsn, connect_timeout=10)
        except _TRANSIENT_DB_ERRORS as exc:
            attempt += 1
            delay = min(cfg.retry_backoff_base ** attempt, 30.0)
            _capture(exc)
            log.warning("db connect failed; backing off",
                        extra={"attempt": attempt, "delay_s": round(delay, 2), "error": str(exc)})
            if not _interruptible_sleep(delay, shutdown):
                return None
    return None


def _interruptible_sleep(seconds: float, shutdown: ShutdownFlag) -> bool:
    """Sleep in small slices so a shutdown signal is honored promptly.
    Returns False if shutdown was requested during the sleep."""
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        if shutdown.requested:
            return False
        time.sleep(max(0.0, min(0.5, deadline - time.monotonic())))
    return not shutdown.requested


def main() -> None:
    cfg = Config.from_env()
    _init_sentry(cfg)
    shutdown = ShutdownFlag()
    signal.signal(signal.SIGTERM, shutdown.request)
    signal.signal(signal.SIGINT, shutdown.request)

    s3 = s3_client(cfg)
    log.info(
        "worker started",
        extra={"worker_id": WORKER_ID, "poll_interval_s": cfg.poll_interval,
               "parser_version": cfg.parser_version, "stuck_timeout_s": cfg.stuck_timeout_seconds},
    )

    while not shutdown.requested:
        conn = _connect_with_backoff(cfg, shutdown)
        if conn is None:
            break
        try:
            reset_stuck_jobs(conn, cfg)
            # Drain the queue, but stop between jobs if asked to shut down.
            while not shutdown.requested and process_one(conn, s3, cfg):
                pass
        except _TRANSIENT_DB_ERRORS as exc:
            _capture(exc)
            log.warning("db connection lost; reconnecting", extra={"error": str(exc)})
            _safe_close(conn)
            continue  # reconnect immediately
        except Exception as exc:
            _capture(exc)
            log.exception("worker loop error")
        finally:
            _safe_close(conn)
        if shutdown.requested:
            break
        _interruptible_sleep(cfg.poll_interval, shutdown)

    log.info("worker stopped cleanly", extra={"worker_id": WORKER_ID})


if __name__ == "__main__":
    main()
