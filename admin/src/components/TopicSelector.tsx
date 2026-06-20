import { useEffect, useState } from "react";
import { useChapters, useSubjects, useTopics } from "@/hooks/queries";
import { Select } from "@/components/ui/select";
import { Label } from "@/components/ui/label";

interface TopicSelectorProps {
  value: string;
  onChange: (topicId: string) => void;
}

export function TopicSelector({ value, onChange }: TopicSelectorProps) {
  const [subjectId, setSubjectId] = useState("");
  const [chapterId, setChapterId] = useState("");

  const subjects = useSubjects();
  const chapters = useChapters(subjectId || undefined);
  const topics = useTopics(chapterId || undefined);

  useEffect(() => {
    setChapterId("");
    onChange("");
  }, [subjectId, onChange]);

  useEffect(() => {
    onChange("");
  }, [chapterId, onChange]);

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
      <div>
        <Label>Môn</Label>
        <Select value={subjectId} onChange={(e) => setSubjectId(e.target.value)}>
          <option value="">— chọn môn —</option>
          {(subjects.data ?? []).map((s) => (
            <option key={s.id} value={s.id}>
              {s.name}
            </option>
          ))}
        </Select>
      </div>
      <div>
        <Label>Chương</Label>
        <Select
          value={chapterId}
          onChange={(e) => setChapterId(e.target.value)}
          disabled={!subjectId}
        >
          <option value="">— chọn chương —</option>
          {(chapters.data ?? []).map((c) => (
            <option key={c.id} value={c.id}>
              {c.title}
            </option>
          ))}
        </Select>
      </div>
      <div>
        <Label>Chủ đề</Label>
        <Select value={value} onChange={(e) => onChange(e.target.value)} disabled={!chapterId}>
          <option value="">— chọn chủ đề —</option>
          {(topics.data ?? []).map((t) => (
            <option key={t.id} value={t.id}>
              {t.title}
            </option>
          ))}
        </Select>
      </div>
    </div>
  );
}
