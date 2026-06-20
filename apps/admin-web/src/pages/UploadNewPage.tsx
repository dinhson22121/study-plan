import { useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/toast";
import { sha256Hex } from "@/lib/checksum";
import { env } from "@/lib/env";
import { formatBytes } from "@/lib/utils";
import * as uploads from "@/api/uploads";
import { ApiError } from "@/api/client";

type Phase = "idle" | "checksum" | "uploading" | "completing";

export function UploadNewPage() {
  const navigate = useNavigate();
  const { toast } = useToast();
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [phase, setPhase] = useState<Phase>("idle");
  const [error, setError] = useState<string | null>(null);

  function pick(f: File | null) {
    setError(null);
    const maxSize = env.uploadMaxFileSizeBytes;
    if (!f) {
      setFile(null);
      return;
    }
    if (f.type !== "application/pdf") {
      setError("Chỉ chấp nhận file PDF.");
      return;
    }
    if (f.size > maxSize) {
      setError(`File vượt quá ${formatBytes(maxSize)}.`);
      return;
    }
    setFile(f);
  }

  async function submit() {
    if (!file) return;
    setError(null);
    try {
      setPhase("checksum");
      const checksum = await sha256Hex(file);

      const init = await uploads.initUpload({
        filename: file.name,
        content_type: "application/pdf",
        file_size: file.size,
        checksum_sha256: checksum,
      });

      setPhase("uploading");
      await uploads.putToPresignedUrl(init.upload_url, init.method, init.headers, file);

      setPhase("completing");
      const done = await uploads.completeUpload(init.asset_id);

      toast("Upload thành công", "success");
      navigate(`/uploads/${done.asset.id}`);
    } catch (err) {
      setPhase("idle");
      setError(err instanceof ApiError || err instanceof Error ? err.message : "Upload thất bại");
    }
  }

  const busy = phase !== "idle";
  const phaseLabel: Record<Phase, string> = {
    idle: "Upload",
    checksum: "Đang tính checksum...",
    uploading: "Đang tải lên storage...",
    completing: "Đang hoàn tất...",
  };

  return (
    <div className="mx-auto max-w-xl space-y-4">
      <h1 className="text-2xl font-bold text-slate-800">Upload PDF đề bài</h1>
      <Card>
        <CardHeader>
          <CardTitle>Chọn file</CardTitle>
        </CardHeader>
        <CardBody className="space-y-4">
          {error && (
            <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{error}</div>
          )}
          <div>
            <Label>File PDF (tối đa {formatBytes(env.uploadMaxFileSizeBytes)})</Label>
            <input
              ref={inputRef}
              type="file"
              accept="application/pdf"
              disabled={busy}
              onChange={(e) => pick(e.target.files?.[0] ?? null)}
              className="block w-full text-sm text-slate-600 file:mr-3 file:rounded-md file:border-0 file:bg-blue-600 file:px-4 file:py-2 file:text-white"
            />
          </div>
          {file && (
            <div className="rounded-md bg-slate-50 px-3 py-2 text-sm text-slate-600">
              {file.name} · {formatBytes(file.size)}
            </div>
          )}
          <Button onClick={submit} disabled={!file || busy} className="w-full">
            {phaseLabel[phase]}
          </Button>
        </CardBody>
      </Card>
    </div>
  );
}
