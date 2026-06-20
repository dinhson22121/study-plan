export interface Envelope<T> {
  success: boolean;
  data: T;
  meta?: Meta;
  error?: ApiErrorBody;
}

export interface Meta {
  total: number;
  page: number;
  limit: number;
}

export interface ApiErrorBody {
  code: string;
  message: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
}

export type AssetStatus = "PENDING" | "UPLOADED" | "VERIFIED" | "DELETED" | "FAILED";
export type EntityType = "QUESTION" | "EXAM" | "CONTENT" | "ATTACHMENT" | "";

export interface Asset {
  id: string;
  object_key: string;
  bucket_name: string;
  original_filename: string;
  content_type: string;
  file_size: number;
  checksum_sha256: string;
  status: AssetStatus;
  uploaded_by: string;
  entity_type: EntityType;
  entity_id: string;
  storage_provider: string;
  created_at: string;
  verified_at: string | null;
  deleted_at: string | null;
}

export type ParseJobStatus = "QUEUED" | "PROCESSING" | "PARSED" | "REVIEW_REQUIRED" | "FAILED";

export interface ParseJob {
  id: string;
  asset_id: string;
  status: ParseJobStatus;
  parser_version: string;
  attempt_count: number;
  error_message: string;
  claimed_at: string | null;
  started_at: string | null;
  finished_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface InitUploadResponse {
  asset_id: string;
  object_key: string;
  upload_url: string;
  method: string;
  headers: Record<string, string>;
  expires_at: string;
}

export interface CompleteUploadResponse {
  asset: Asset;
  parse_job_id: string;
}

export type DraftStatus = "DRAFT" | "PUBLISHED";

export interface QuestionDraftOption {
  id: string;
  question_draft_id: string;
  option_label: string;
  option_text: string;
  is_correct_inferred: boolean;
  order_index: number;
}

export interface QuestionDraft {
  id: string;
  asset_id: string;
  parse_job_id: string;
  question_number: number;
  question_type: string;
  stem: string;
  explanation_raw: string;
  answer_key_raw: string;
  parse_confidence: number;
  status: DraftStatus;
  reviewed_by: string;
  reviewed_at: string | null;
  published_question_id: string;
  options: QuestionDraftOption[];
  created_at: string;
  updated_at: string;
}

export interface Subject {
  id: string;
  code: string;
  name: string;
  grade_level: number;
}

export interface Chapter {
  id: string;
  subject_id: string;
  title: string;
  order_index: number;
}

export interface Topic {
  id: string;
  chapter_id: string;
  title: string;
  order_index: number;
}

export type Difficulty = "EASY" | "MEDIUM" | "HARD";

export interface QuestionSummary {
  id: string;
  topic_id: string;
  type: string;
  stem: string;
  difficulty: Difficulty;
}
