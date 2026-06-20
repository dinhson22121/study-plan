import { useEffect, useState, type ReactNode } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useAsset, useParseJobs, useQuestions } from "@/hooks/queries";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { Dialog } from "@/components/ui/dialog";
import { Select } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/toast";
import { AssetStatusBadge, ParseStatusBadge } from "@/components/AssetStatusBadge";
import { formatBytes, formatDate } from "@/lib/utils";
import { TopicSelector } from "@/components/TopicSelector";
import * as uploads from "@/api/uploads";

const ENTITY_TYPES = ["QUESTION", "EXAM", "CONTENT", "ATTACHMENT"];

export function UploadDetailPage() {
  const { assetId } = useParams<{ assetId: string }>();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const { toast } = useToast();
  const asset = useAsset(assetId);
  const jobs = useParseJobs(assetId);
  const [linkOpen, setLinkOpen] = useState(false);

  const retry = useMutation({
    mutationFn: () => uploads.retryParse(assetId as string),
    onSuccess: () => {
      toast("Đã tạo lại parse job", "success");
      qc.invalidateQueries({ queryKey: ["parse-jobs", assetId] });
    },
    onError: (e: Error) => toast(e.message, "error"),
  });

  const remove = useMutation({
    mutationFn: () => uploads.deleteAsset(assetId as string),
    onSuccess: () => {
      toast("Đã xoá asset", "success");
      navigate("/uploads");
    },
    onError: (e: Error) => toast(e.message, "error"),
  });

  if (asset.isLoading) return <Spinner label="Đang tải..." />;
  if (asset.isError || !asset.data) return <p className="text-red-600">Không tìm thấy asset.</p>;

  const a = asset.data;

  return (
    <div className="space-y-5">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-800">{a.original_filename}</h1>
        <div className="flex gap-2">
          <Link to={`/uploads/${a.id}/drafts`}>
            <Button>Review & Publish</Button>
          </Link>
          <Button variant="secondary" onClick={() => setLinkOpen(true)}>
            Link entity
          </Button>
          <Button
            variant="danger"
            onClick={() => {
              if (confirm("Xoá asset này?")) remove.mutate();
            }}
            disabled={remove.isPending}
          >
            Xoá
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Thông tin asset</CardTitle>
        </CardHeader>
        <CardBody>
          <dl className="grid grid-cols-1 gap-x-8 gap-y-2 text-sm sm:grid-cols-2">
            <Row label="Trạng thái">
              <AssetStatusBadge status={a.status} />
            </Row>
            <Row label="Dung lượng">{formatBytes(a.file_size)}</Row>
            <Row label="Content type">{a.content_type}</Row>
            <Row label="Checksum">
              <span className="break-all font-mono text-xs">{a.checksum_sha256 || "—"}</span>
            </Row>
            <Row label="Tạo lúc">{formatDate(a.created_at)}</Row>
            <Row label="Verify lúc">{formatDate(a.verified_at ?? undefined)}</Row>
            <Row label="Entity">
              {a.entity_type ? `${a.entity_type} · ${a.entity_id}` : "Chưa link"}
            </Row>
            <Row label="Object key">
              <span className="break-all font-mono text-xs">{a.object_key}</span>
            </Row>
          </dl>
        </CardBody>
      </Card>

      <Card>
        <CardHeader className="flex items-center justify-between">
          <CardTitle>Parse jobs</CardTitle>
          <Button size="sm" variant="secondary" onClick={() => retry.mutate()} disabled={retry.isPending}>
            Parse lại
          </Button>
        </CardHeader>
        <CardBody>
          {jobs.isLoading ? (
            <Spinner label="Đang tải..." />
          ) : (jobs.data ?? []).length === 0 ? (
            <p className="text-sm text-slate-500">Chưa có parse job.</p>
          ) : (
            <ul className="space-y-2">
              {(jobs.data ?? []).map((j) => (
                <li
                  key={j.id}
                  className="flex items-center justify-between rounded-md border border-slate-100 px-3 py-2"
                >
                  <div className="flex items-center gap-3">
                    <ParseStatusBadge status={j.status} />
                    <span className="text-xs text-slate-500">lần {j.attempt_count}</span>
                    {j.status === "FAILED" && j.error_message && (
                      <span className="text-xs text-red-600">{j.error_message}</span>
                    )}
                  </div>
                  <span className="text-xs text-slate-400">{formatDate(j.created_at)}</span>
                </li>
              ))}
            </ul>
          )}
        </CardBody>
      </Card>

      <LinkEntityDialog
        open={linkOpen}
        assetId={a.id}
        onClose={() => setLinkOpen(false)}
        onLinked={() => {
          setLinkOpen(false);
          qc.invalidateQueries({ queryKey: ["asset", assetId] });
        }}
      />
    </div>
  );
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex justify-between gap-4 border-b border-slate-50 py-1">
      <dt className="text-slate-500">{label}</dt>
      <dd className="text-right text-slate-700">{children}</dd>
    </div>
  );
}

function LinkEntityDialog({
  open,
  assetId,
  onClose,
  onLinked,
}: {
  open: boolean;
  assetId: string;
  onClose: () => void;
  onLinked: () => void;
}) {
  const { toast } = useToast();
  const [entityType, setEntityType] = useState(ENTITY_TYPES[0]);
  const [entityId, setEntityId] = useState("");
  const [questionTopicId, setQuestionTopicId] = useState("");
  const questionOptions = useQuestions(entityType === "QUESTION" ? questionTopicId : undefined);

  useEffect(() => {
    setEntityId("");
    setQuestionTopicId("");
  }, [entityType]);

  const link = useMutation({
    mutationFn: () => uploads.linkEntity(assetId, entityType, entityId),
    onSuccess: () => {
      toast("Đã link entity", "success");
      onLinked();
    },
    onError: (e: Error) => toast(e.message, "error"),
  });

  return (
    <Dialog
      open={open}
      title="Link asset vào entity"
      onClose={onClose}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Huỷ
          </Button>
          <Button onClick={() => link.mutate()} disabled={!entityId || link.isPending}>
            Link
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        <div>
          <Label>Entity type</Label>
          <Select value={entityType} onChange={(e) => setEntityType(e.target.value)}>
            {ENTITY_TYPES.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </Select>
        </div>
        {entityType === "QUESTION" ? (
          <>
            <TopicSelector value={questionTopicId} onChange={setQuestionTopicId} />
            <div>
              <Label>Question</Label>
              <Select
                value={entityId}
                onChange={(e) => setEntityId(e.target.value)}
                disabled={!questionTopicId || questionOptions.isLoading}
              >
                <option value="">— chọn question —</option>
                {(questionOptions.data ?? []).map((q) => (
                  <option key={q.id} value={q.id}>
                    {q.stem}
                  </option>
                ))}
              </Select>
            </div>
          </>
        ) : (
          <div>
            <Label>Entity ID</Label>
            <Input value={entityId} onChange={(e) => setEntityId(e.target.value)} placeholder="UUID" />
          </div>
        )}
      </div>
    </Dialog>
  );
}
