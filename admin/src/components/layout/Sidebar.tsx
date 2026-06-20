import { NavLink } from "react-router-dom";
import { cn } from "@/lib/utils";
import { env } from "@/lib/env";

const links = [
  { to: "/dashboard", label: "Dashboard" },
  { to: "/uploads", label: "Uploads" },
  { to: "/uploads/new", label: "Upload PDF" },
];

export function Sidebar() {
  return (
    <aside className="w-56 shrink-0 border-r border-slate-200 bg-white">
      <div className="px-5 py-4 text-lg font-bold text-blue-700">{env.appName}</div>
      <nav className="flex flex-col gap-1 px-3">
        {links.map((l) => (
          <NavLink
            key={l.to}
            to={l.to}
            end={l.to === "/uploads"}
            className={({ isActive }) =>
              cn(
                "rounded-md px-3 py-2 text-sm font-medium",
                isActive ? "bg-blue-50 text-blue-700" : "text-slate-600 hover:bg-slate-100",
              )
            }
          >
            {l.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
