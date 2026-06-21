import { api, unwrap } from "./client";
import type { Envelope, QuestionSummary } from "./types";

export async function listQuestions(topicId: string, limit = 100): Promise<QuestionSummary[]> {
  const data = await unwrap<QuestionSummary[] | null>(
    api.get<Envelope<QuestionSummary[] | null>>("/questions", {
      params: { topic_id: topicId, limit },
    }),
  );
  return data ?? [];
}
