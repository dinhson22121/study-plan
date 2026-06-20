from __future__ import annotations

import dataclasses
import logging
import os
import re
import time
import uuid

import boto3
import fitz
import psycopg
from botocore.config import Config as BotoConfig

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s pdf_parser %(message)s",
)
log = logging.getLogger("pdf_parser")


@dataclasses.dataclass
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
        )


def s3_client(cfg: Config):
    kwargs = {
        "region_name": cfg.s3_region,
        "aws_access_key_id": cfg.s3_access_key,
        "aws_secret_access_key": cfg.s3_secret_key,
        "config": BotoConfig(s3={"addressing_style": "path" if cfg.s3_path_style else "auto"}),
    }
    if cfg.s3_endpoint:
        kwargs["endpoint_url"] = cfg.s3_endpoint
    return boto3.client("s3", **kwargs)


WORKER_ID = f"pdf-parser-{uuid.uuid4().hex[:8]}"


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
        RETURNING p.id, p.asset_id
    """
    with conn.cursor() as cur:
        cur.execute(sql, (WORKER_ID, parser_version))
        row = cur.fetchone()
    conn.commit()
    return row


def fetch_asset(conn, asset_id):
    with conn.cursor() as cur:
        cur.execute(
            "SELECT object_key, bucket_name FROM uploaded_asset WHERE id = %s", (asset_id,)
        )
        return cur.fetchone()


def finish_job(conn, job_id, status, raw_text="", error=""):
    with conn.cursor() as cur:
        cur.execute(
            """UPDATE parse_job
               SET status=%s, finished_at=now(), updated_at=now(),
                   raw_text=%s, error_message=%s
               WHERE id=%s""",
            (status, raw_text[:200000], error[:2000], job_id),
        )
    conn.commit()


QUESTION_RE = re.compile(r"(?:^|\n)\s*(?:Câu|Bài|Question)\s*(\d+)\s*[:.)]?\s*", re.IGNORECASE)
OPTION_RE = re.compile(r"(?:^|\n)\s*([A-D])\s*[.)]\s*(.+)")
ANSWER_RE = re.compile(r"(?:Đáp\s*án|Answer|ĐA)\s*[:.]?\s*([A-D])", re.IGNORECASE)


@dataclasses.dataclass
class ParsedQuestion:
    number: int
    stem: str
    options: list
    confidence: float


def extract_text(pdf_bytes: bytes) -> str:
    doc = fitz.open(stream=pdf_bytes, filetype="pdf")
    try:
        return "\n".join(page.get_text() for page in doc)
    finally:
        doc.close()


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


def save_drafts(conn, job_id, asset_id, questions: list):
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


def process_one(conn, s3, cfg: Config) -> bool:
    claimed = claim_job(conn, cfg.parser_version)
    if not claimed:
        return False
    job_id, asset_id = claimed
    log.info("claimed job %s (asset %s)", job_id, asset_id)

    try:
        asset = fetch_asset(conn, asset_id)
        if not asset:
            finish_job(conn, job_id, "FAILED", error="asset not found")
            return True
        object_key, bucket = asset
        obj = s3.get_object(Bucket=bucket or cfg.s3_bucket, Key=object_key)
        pdf_bytes = obj["Body"].read()

        text = extract_text(pdf_bytes)
        questions = parse_questions(text)
        save_drafts(conn, job_id, asset_id, questions)

        status = "REVIEW_REQUIRED" if needs_review(questions) else "PARSED"
        finish_job(conn, job_id, status, raw_text=text)
        log.info("job %s -> %s (%d drafts)", job_id, status, len(questions))
    except Exception as exc:
        conn.rollback()
        log.exception("job %s failed", job_id)
        finish_job(conn, job_id, "FAILED", error=str(exc))
    return True


def main():
    cfg = Config.from_env()
    s3 = s3_client(cfg)
    log.info("worker %s started; polling every %.1fs", WORKER_ID, cfg.poll_interval)
    while True:
        try:
            with psycopg.connect(cfg.dsn) as conn:
                while process_one(conn, s3, cfg):
                    pass
        except Exception:
            log.exception("worker loop error")
        time.sleep(cfg.poll_interval)


if __name__ == "__main__":
    main()
