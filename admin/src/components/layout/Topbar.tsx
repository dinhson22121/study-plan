import { useNavigate } from "react-router-dom";
import { useAuth } from "@/auth/AuthContext";
import { Button } from "@/components/ui/button";

export function Topbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    await logout();
    navigate("/login", { replace: true });
  }

  return (
    <header className="flex items-center justify-between border-b border-slate-200 bg-white px-6 py-3">
      <div className="text-sm text-slate-500">Admin Console</div>
      <div className="flex items-center gap-3">
        <span className="text-sm text-slate-600">{user?.id?.slice(0, 8)} · {user?.role}</span>
        <Button variant="secondary" size="sm" onClick={handleLogout} data-testid="logout">
          Đăng xuất
        </Button>
      </div>
    </header>
  );
}
