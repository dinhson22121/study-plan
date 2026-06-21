import { api, unwrap } from "./client";
import type { Envelope, TokenPair } from "./types";

export function login(email: string, password: string): Promise<TokenPair> {
  return unwrap<TokenPair>(api.post<Envelope<TokenPair>>("/auth/login", { email, password }));
}

export async function logout(refreshToken: string): Promise<void> {
  await api.post("/auth/logout", { refresh_token: refreshToken }).catch(() => undefined);
}
