import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

type Tone = "gray" | "blue" | "green" | "amber" | "red" | "violet";

const tones: Record<Tone, string> = {
  gray: "bg-slate-100 text-slate-700",
  blue: "bg-blue-100 text-blue-700",
  green: "bg-green-100 text-green-700",
  amber: "bg-amber-100 text-amber-800",
  red: "bg-red-100 text-red-700",
  violet: "bg-violet-100 text-violet-700",
};

export function Badge({ tone = "gray", children }: { tone?: Tone; children: ReactNode }) {
  return (
    <span className={cn("inline-block rounded-full px-2.5 py-0.5 text-xs font-medium", tones[tone])}>
      {children}
    </span>
  );
}
