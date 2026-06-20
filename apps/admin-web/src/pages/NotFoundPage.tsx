import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <p className="text-4xl font-bold text-slate-800">404</p>
        <p className="mt-2 text-slate-500">Không tìm thấy trang.</p>
        <Link to="/dashboard" className="mt-4 inline-block text-blue-600 hover:underline">
          Về Dashboard
        </Link>
      </div>
    </div>
  );
}
