import { Link } from "react-router-dom";
import { useAssets } from "@/hooks/queries";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import { Button } from "@/components/ui/button";
import { AssetStatusBadge } from "@/components/AssetStatusBadge";
import { formatDate } from "@/lib/utils";

export function DashboardPage() {
  const { data, isLoading } = useAssets();
  const assets = data?.items ?? [];
  const needReview = assets.filter((a) => a.status === "UPLOADED").length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-800">Dashboard</h1>
        <Link to="/uploads/new">
          <Button>Upload PDF</Button>
        </Link>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard label="Tổng số asset" value={assets.length} />
        <StatCard label="Đã upload" value={needReview} />
        <StatCard label="Hiển thị gần đây" value={Math.min(assets.length, 5)} />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Asset gần đây</CardTitle>
        </CardHeader>
        <CardBody>
          {isLoading ? (
            <Spinner label="Đang tải..." />
          ) : assets.length === 0 ? (
            <p className="text-sm text-slate-500">Chưa có asset nào.</p>
          ) : (
            <ul className="divide-y divide-slate-100">
              {assets.slice(0, 5).map((a) => (
                <li key={a.id} className="flex items-center justify-between py-2">
                  <Link to={`/uploads/${a.id}`} className="text-sm text-blue-600 hover:underline">
                    {a.original_filename}
                  </Link>
                  <div className="flex items-center gap-3">
                    <AssetStatusBadge status={a.status} />
                    <span className="text-xs text-slate-400">{formatDate(a.created_at)}</span>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </CardBody>
      </Card>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <Card>
      <CardBody>
        <p className="text-sm text-slate-500">{label}</p>
        <p className="mt-1 text-3xl font-bold text-slate-800">{value}</p>
      </CardBody>
    </Card>
  );
}
