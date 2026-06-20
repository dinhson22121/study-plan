import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { tokenStore } from "./tokenStore";
import { decodeJwt } from "@/lib/jwt";
import { setOnUnauthorized } from "@/api/client";
import * as authApi from "@/api/auth";

export interface AuthUser {
  id: string;
  role: string;
}

interface AuthContextValue {
  user: AuthUser | null;
  isAuthenticated: boolean;
  isAdmin: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function userFromToken(token: string | null): AuthUser | null {
  if (!token) return null;
  const claims = decodeJwt(token);
  if (!claims) return null;
  if (claims.exp && claims.exp * 1000 < Date.now()) return null;
  return { id: claims.sub, role: claims.role };
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(() => userFromToken(tokenStore.access));

  useEffect(() => {
    setOnUnauthorized(() => setUser(null));
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: user !== null,
      isAdmin: user?.role === "ADMIN",
      async login(email, password) {
        const pair = await authApi.login(email, password);
        tokenStore.set(pair.access_token, pair.refresh_token);
        setUser(userFromToken(pair.access_token));
      },
      async logout() {
        const rt = tokenStore.refresh;
        if (rt) await authApi.logout(rt);
        tokenStore.clear();
        setUser(null);
      },
    }),
    [user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
