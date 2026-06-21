import { useState } from "react";
import { useParams } from "react-router-dom";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useDrafts } from "@/hooks/queries";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { Input, Textarea } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/components/ui/toast";
import { TopicSelector } from "@/components/TopicSelector";
import * as drafts from "@/api/drafts";
import type { Difficulty, QuestionDraft } from "@/api/types";

const DIFFICULTIES: Difficulty[] = ["EASY", "MEDIUM", "HARD"];

export function DraftReviewPage() {
  const { assetId } = useParams<{ assetId: string }>();
  const qc = useQueryClient();
  const { toast } = useToast();
  const { data, isLoading } = useDrafts(assetId);
  const [topicId, setTopicId] = useState("");
  const [difficulty, setDifficulty] = useState<Difficulty>("MEDIUM");

  const publishAll = useMutation({
    mutationFn: () => drafts.publishByAsset(assetId as string, topicId, difficulty),
    onSuccess: () => {
      toast("Đã publish toàn bộ draft", "success");
      qc.invalidateQueries({ queryKey: ["drafts", assetId] });
    },
    onError: (e: Error) => toast(e.message, "error"),
  });

  const items = data ?? [];
  const pending = items.filter((d) => d.status !== "PUBLISHED").length;

  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-bold text-slate-800">Review draft questions</h1>

      <Card>
        <CardHeader>
          <CardTitle>Chọn đích publish</CardTitle>
        </CardHeader>
        <CardBody className="space-y-4">
          <TopicSelector value={topicId} onChange={setTopicId} />
          <div className="flex items-end gap-4">
            <div className="w-40">
              <Label>Độ khó</Label>
              <Select value={difficulty} onChange={(e) => setDifficulty(e.target.value as Difficulty)}>
                {DIFFICULTIES.map((d) => (
                  <option key={d} value={d}>
                    {d}
                  </option>
                ))}
              </Select>
            </div>
            <Button
              onClick={() => {
                if (!topicId) {
                  toast("Hãy chọn chủ đề trước", "error");
                  return;
                }
                if (confirm(`Publish ${pending} draft chưa publish?`)) publishAll.mutate();
              }}
              disabled={pending === 0 || publishAll.isPending}
            >
              Publish tất cả ({pending})
            </Button>
          </div>
        </CardBody>
      </Card>

      {isLoading ? (
        <Spinner label="Đang tải draft..." />
      ) : items.length === 0 ? (
        <p className="text-sm text-slate-500">Chưa có draft. Hãy chờ worker parse xong.</p>
      ) : (
        <div className="space-y-4">
          {items.map((d) => (
            <DraftCard
              key={d.id}
              draft={d}
              assetId={assetId as string}
              topicId={topicId}
              difficulty={difficulty}
            />
          ))}
        </div>
      )}
    </div>
  );
}

interface OptionState {
  id: string;
  label: string;
  text: string;
}

function DraftCard({
  draft,
  assetId,
  topicId,
  difficulty,
}: {
  draft: QuestionDraft;
  assetId: string;
  topicId: string;
  difficulty: Difficulty;
}) {
  const qc = useQueryClient();
  const { toast } = useToast();
  const published = draft.status === "PUBLISHED";

  const [stem, setStem] = useState(draft.stem);
  const [explanation, setExplanation] = useState(draft.explanation_raw);
  const [options, setOptions] = useState<OptionState[]>(
    draft.options.map((o) => ({ id: o.id, label: o.option_label, text: o.option_text })),
  );
  const [correctId, setCorrectId] = useState<string>(
    draft.options.find((o) => o.is_correct_inferred)?.id ?? "",
  );

  const save = useMutation({
    mutationFn: async () => {
      await drafts.updateDraft(draft.id, stem, explanation);
      await Promise.all(
        options.map((o) => drafts.updateOption(draft.id, o.id, o.text, o.id === correctId)),
      );
    },
    onSuccess: () => {
      toast("Đã lưu draft", "success");
      qc.invalidateQueries({ queryKey: ["drafts", assetId] });
    },
    onError: (e: Error) => toast(e.message, "error"),
  });

  const publish = useMutation({
    mutationFn: () => drafts.publishDraft(draft.id, topicId, difficulty),
    onSuccess: () => {
      toast("Đã publish draft", "success");
      qc.invalidateQueries({ queryKey: ["drafts", assetId] });
    },
    onError: (e: Error) => toast(e.message, "error"),
  });

  const missingCorrect = correctId === "";

  return (
    <Card className={missingCorrect && !published ? "border-amber-300" : undefined}>
      <CardHeader className="flex items-center justify-between">
        <CardTitle>
          Câu {draft.question_number}{" "}
          <span className="ml-2 text-xs font-normal text-slate-400">
            confidence {Math.round(draft.parse_confidence * 100)}%
          </span>
        </CardTitle>
        <div className="flex items-center gap-2">
          {published ? (
            <Badge tone="green">PUBLISHED</Badge>
          ) : missingCorrect ? (
            <Badge tone="amber">THIẾU ĐÁP ÁN</Badge>
          ) : (
            <Badge tone="gray">DRAFT</Badge>
          )}
        </div>
      </CardHeader>
      <CardBody className="space-y-3">
        <div>
          <Label>Đề bài</Label>
          <Textarea rows={2} value={stem} disabled={published} onChange={(e) => setStem(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Đáp án (chọn 1 đáp án đúng)</Label>
          {options.map((o, idx) => (
            <div key={o.id} className="flex items-center gap-2">
              <input
                type="radio"
                name={`correct-${draft.id}`}
                checked={correctId === o.id}
                disabled={published}
                onChange={() => setCorrectId(o.id)}
              />
              <span className="w-5 text-sm font-medium text-slate-500">{o.label}</span>
              <Input
                value={o.text}
                disabled={published}
                onChange={(e) => {
                  const next = [...options];
                  next[idx] = { ...o, text: e.target.value };
                  setOptions(next);
                }}
              />
            </div>
          ))}
        </div>
        <div>
          <Label>Giải thích</Label>
          <Textarea
            rows={2}
            value={explanation}
            disabled={published}
            onChange={(e) => setExplanation(e.target.value)}
          />
        </div>
        {!published && (
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => save.mutate()} disabled={save.isPending}>
              Lưu
            </Button>
            <Button
              onClick={() => {
                if (!topicId) {
                  toast("Hãy chọn chủ đề trước", "error");
                  return;
                }
                publish.mutate();
              }}
              disabled={publish.isPending}
            >
              Publish
            </Button>
          </div>
        )}
      </CardBody>
    </Card>
  );
}
