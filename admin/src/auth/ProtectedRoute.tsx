import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "./AuthContext";

export function ProtectedRoute() {
  const { isAuthenticated, isAdmin } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (!isAdmin) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <p className="text-4xl font-bold text-slate-800">403</p>
          <p className="mt-2 text-slate-500">Tài khoản của bạn không có quyền ADMIN.</p>
        </div>
      </div>
    );
  }
  return <Outlet />;
}
