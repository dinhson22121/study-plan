import { Badge } from "@/components/ui/badge";
import type { AssetStatus, ParseJobStatus } from "@/api/types";

export function AssetStatusBadge({ status }: { status: AssetStatus }) {
  const tone =
    status === "UPLOADED" || status === "VERIFIED"
      ? "green"
      : status === "PENDING"
        ? "amber"
        : status === "FAILED"
          ? "red"
          : "gray";
  return <Badge tone={tone}>{status}</Badge>;
}

export function ParseStatusBadge({ status }: { status: ParseJobStatus }) {
  const tone =
    status === "PARSED"
      ? "green"
      : status === "REVIEW_REQUIRED"
        ? "violet"
        : status === "FAILED"
          ? "red"
          : status === "PROCESSING"
            ? "blue"
            : "amber";
  return <Badge tone={tone}>{status}</Badge>;
}
