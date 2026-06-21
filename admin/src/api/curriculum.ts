import { api, unwrap } from "./client";
import type { Chapter, Envelope, Subject, Topic } from "./types";

export async function listSubjects(): Promise<Subject[]> {
  const data = await unwrap<Subject[] | null>(api.get<Envelope<Subject[] | null>>("/curriculum/subjects"));
  return data ?? [];
}

export async function listChapters(subjectId: string): Promise<Chapter[]> {
  const data = await unwrap<Chapter[] | null>(
    api.get<Envelope<Chapter[] | null>>(`/curriculum/subjects/${subjectId}/chapters`),
  );
  return data ?? [];
}

export async function listTopics(chapterId: string): Promise<Topic[]> {
  const data = await unwrap<Topic[] | null>(
    api.get<Envelope<Topic[] | null>>(`/curriculum/chapters/${chapterId}/topics`),
  );
  return data ?? [];
}
