import { api, unwrap, unwrapList } from "./client";
import type {
  Asset,
  CompleteUploadResponse,
  Envelope,
  InitUploadResponse,
  Meta,
  ParseJob,
} from "./types";

export interface InitUploadBody {
  filename: string;
  content_type: string;
  file_size: number;
  checksum_sha256?: string;
}

export function initUpload(body: InitUploadBody): Promise<InitUploadResponse> {
  return unwrap<InitUploadResponse>(
    api.post<Envelope<InitUploadResponse>>("/admin/uploads/init", body),
  );
}

export function completeUpload(assetId: string): Promise<CompleteUploadResponse> {
  return unwrap<CompleteUploadResponse>(
    api.post<Envelope<CompleteUploadResponse>>("/admin/uploads/complete", { asset_id: assetId }),
  );
}

export async function listAssets(params: {
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<{ items: Asset[]; meta?: Meta }> {
  const { data, meta } = await unwrapList<Asset[] | null>(
    api.get<Envelope<Asset[] | null>>("/admin/uploads", { params }),
  );
  return { items: data ?? [], meta };
}

export function getAsset(id: string): Promise<Asset> {
  return unwrap<Asset>(api.get<Envelope<Asset>>(`/admin/uploads/${id}`));
}

export async function listParseJobs(id: string): Promise<ParseJob[]> {
  const jobs = await unwrap<ParseJob[] | null>(
    api.get<Envelope<ParseJob[] | null>>(`/admin/uploads/${id}/parse-jobs`),
  );
  return jobs ?? [];
}

export function retryParse(id: string): Promise<ParseJob> {
  return unwrap<ParseJob>(api.post<Envelope<ParseJob>>(`/admin/uploads/${id}/parse`));
}

export function linkEntity(id: string, entityType: string, entityId: string): Promise<unknown> {
  return unwrap<unknown>(
    api.post<Envelope<unknown>>(`/admin/uploads/${id}/link`, {
      entity_type: entityType,
      entity_id: entityId,
    }),
  );
}

export function deleteAsset(id: string): Promise<unknown> {
  return unwrap<unknown>(api.delete<Envelope<unknown>>(`/admin/uploads/${id}`));
}

export async function putToPresignedUrl(
  url: string,
  method: string,
  headers: Record<string, string>,
  file: File,
): Promise<void> {
  const res = await fetch(url, { method: method || "PUT", headers, body: file });
  if (!res.ok) {
    throw new Error(`upload to storage failed: ${res.status} ${res.statusText}`);
  }
}
