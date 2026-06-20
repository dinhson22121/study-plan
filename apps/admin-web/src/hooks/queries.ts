import { useQuery } from "@tanstack/react-query";
import { env } from "@/lib/env";
import * as uploads from "@/api/uploads";
import * as drafts from "@/api/drafts";
import * as curriculum from "@/api/curriculum";
import * as questions from "@/api/questions";
import type { ParseJob } from "@/api/types";

export interface UseAssetsParams {
  status?: string;
  limit?: number;
  offset?: number;
}

export function useAssets(params?: UseAssetsParams) {
  const status = params?.status;
  const limit = params?.limit ?? 20;
  const offset = params?.offset ?? 0;
  return useQuery({
    queryKey: ["assets", status ?? "all", limit, offset],
    queryFn: () => uploads.listAssets({ status, limit, offset }),
  });
}

export function useAsset(id: string | undefined) {
  return useQuery({
    queryKey: ["asset", id],
    queryFn: () => uploads.getAsset(id as string),
    enabled: Boolean(id),
  });
}

export function useParseJobs(id: string | undefined) {
  return useQuery({
    queryKey: ["parse-jobs", id],
    queryFn: () => uploads.listParseJobs(id as string),
    enabled: Boolean(id),
    refetchInterval: (query) => {
      const data = query.state.data as ParseJob[] | undefined;
      const latest = data?.[0];
      if (latest && (latest.status === "QUEUED" || latest.status === "PROCESSING")) {
        return env.pollIntervalMs;
      }
      return false;
    },
  });
}

export function useDrafts(assetId: string | undefined) {
  return useQuery({
    queryKey: ["drafts", assetId],
    queryFn: () => drafts.listDraftsByAsset(assetId as string),
    enabled: Boolean(assetId),
  });
}

export function useSubjects() {
  return useQuery({ queryKey: ["subjects"], queryFn: curriculum.listSubjects });
}

export function useChapters(subjectId: string | undefined) {
  return useQuery({
    queryKey: ["chapters", subjectId],
    queryFn: () => curriculum.listChapters(subjectId as string),
    enabled: Boolean(subjectId),
  });
}

export function useTopics(chapterId: string | undefined) {
  return useQuery({
    queryKey: ["topics", chapterId],
    queryFn: () => curriculum.listTopics(chapterId as string),
    enabled: Boolean(chapterId),
  });
}

export function useQuestions(topicId: string | undefined) {
  return useQuery({
    queryKey: ["questions", topicId],
    queryFn: () => questions.listQuestions(topicId as string),
    enabled: Boolean(topicId),
  });
}
