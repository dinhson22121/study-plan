import { useState } from "react";
import { Link } from "react-router-dom";
import { useAssets } from "@/hooks/queries";
import { Card, CardBody } from "@/components/ui/card";
import { Table, TBody, TD, TH, THead, TR } from "@/components/ui/table";
import { Select } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { AssetStatusBadge } from "@/components/AssetStatusBadge";
import { formatBytes, formatDate } from "@/lib/utils";
import type { AssetStatus } from "@/api/types";

const STATUSES: AssetStatus[] = ["PENDING", "UPLOADED", "FAILED", "DELETED"];

export function UploadsPage() {
  const [status, setStatus] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const limit = 20;
  const offset = (page - 1) * limit;
  const { data, isLoading, isError } = useAssets({ status: status || undefined, limit, offset });
  const assets = data?.items ?? [];
  const filtered = assets.filter((a) =>
    a.original_filename.toLowerCase().includes(search.trim().toLowerCase()),
  );
  const total = data?.meta?.total ?? filtered.length;
  const totalPages = Math.max(1, Math.ceil(total / limit));

  function changeStatus(next: string) {
    setStatus(next);
    setPage(1);
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-800">Uploads</h1>
        <Link to="/uploads/new">
          <Button>Upload PDF</Button>
        </Link>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <span className="text-sm text-slate-500">Lọc trạng thái:</span>
          <div className="w-48">
            <Select value={status} onChange={(e) => changeStatus(e.target.value)}>
              <option value="">Tất cả</option>
              {STATUSES.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </Select>
          </div>
        </div>
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Tìm theo tên file trong trang hiện tại..."
          className="w-full rounded-md border border-slate-200 px-3 py-2 text-sm sm:max-w-xs"
        />
      </div>

      <Card>
        <CardBody>
          {isLoading ? (
            <Spinner label="Đang tải..." />
          ) : isError ? (
            <p className="text-sm text-red-600">Không tải được danh sách asset.</p>
          ) : filtered.length === 0 ? (
            <p className="text-sm text-slate-500">Chưa có asset nào.</p>
          ) : (
            <div className="space-y-3">
              <Table>
                <THead>
                  <TR>
                    <TH>Tên file</TH>
                    <TH>Trạng thái</TH>
                    <TH>Dung lượng</TH>
                    <TH>Tạo lúc</TH>
                    <TH></TH>
                  </TR>
                </THead>
                <TBody>
                  {filtered.map((a) => (
                    <TR key={a.id}>
                      <TD className="font-medium text-slate-700">{a.original_filename}</TD>
                      <TD>
                        <AssetStatusBadge status={a.status} />
                      </TD>
                      <TD>{formatBytes(a.file_size)}</TD>
                      <TD className="text-slate-500">{formatDate(a.created_at)}</TD>
                      <TD>
                        <Link to={`/uploads/${a.id}`}>
                          <Button variant="secondary" size="sm">
                            Chi tiết
                          </Button>
                        </Link>
                      </TD>
                    </TR>
                  ))}
                </TBody>
              </Table>
              <div className="flex items-center justify-between text-sm text-slate-500">
                <span>
                  Trang {page}/{totalPages} · tổng {total}
                </span>
                <div className="flex gap-2">
                  <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                    Trước
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => p + 1)}
                  >
                    Sau
                  </Button>
                </div>
              </div>
            </div>
          )}
        </CardBody>
      </Card>
    </div>
  );
}
