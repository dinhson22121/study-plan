import { api, unwrap } from "./client";
import type { Envelope, QuestionDraft } from "./types";

export async function listDraftsByAsset(assetId: string): Promise<QuestionDraft[]> {
  const drafts = await unwrap<QuestionDraft[] | null>(
    api.get<Envelope<QuestionDraft[] | null>>(`/admin/uploads/${assetId}/draft-questions`),
  );
  return drafts ?? [];
}

export function updateDraft(id: string, stem: string, explanation: string): Promise<unknown> {
  return unwrap<unknown>(
    api.put<Envelope<unknown>>(`/admin/question-drafts/${id}`, { stem, explanation }),
  );
}

export function updateOption(
  draftId: string,
  optionId: string,
  text: string,
  isCorrect: boolean,
): Promise<unknown> {
  return unwrap<unknown>(
    api.put<Envelope<unknown>>(`/admin/question-drafts/${draftId}/options/${optionId}`, {
      text,
      is_correct: isCorrect,
    }),
  );
}

export function publishDraft(id: string, topicId: string, difficulty: string): Promise<unknown> {
  return unwrap<unknown>(
    api.post<Envelope<unknown>>(`/admin/question-drafts/${id}/publish`, {
      topic_id: topicId,
      difficulty,
    }),
  );
}

export function publishByAsset(
  assetId: string,
  topicId: string,
  difficulty: string,
): Promise<unknown> {
  return unwrap<unknown>(
    api.post<Envelope<unknown>>(`/admin/uploads/${assetId}/publish`, {
      topic_id: topicId,
      difficulty,
    }),
  );
}
