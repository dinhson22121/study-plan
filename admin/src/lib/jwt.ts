export interface JwtClaims {
  sub: string;
  role: string;
  exp: number;
}

export function decodeJwt(token: string): JwtClaims | null {
  const parts = token.split(".");
  if (parts.length !== 3) return null;
  try {
    const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const json = atob(payload);
    return JSON.parse(json) as JwtClaims;
  } catch {
    return null;
  }
}
