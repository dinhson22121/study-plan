const ACCESS_KEY = "edu_admin_access";
const REFRESH_KEY = "edu_admin_refresh";

let accessToken: string | null = sessionStorage.getItem(ACCESS_KEY);
let refreshToken: string | null = sessionStorage.getItem(REFRESH_KEY);

export const tokenStore = {
  get access(): string | null {
    return accessToken;
  },
  get refresh(): string | null {
    return refreshToken;
  },
  set(access: string, refresh: string): void {
    accessToken = access;
    refreshToken = refresh;
    sessionStorage.setItem(ACCESS_KEY, access);
    sessionStorage.setItem(REFRESH_KEY, refresh);
  },
  clear(): void {
    accessToken = null;
    refreshToken = null;
    sessionStorage.removeItem(ACCESS_KEY);
    sessionStorage.removeItem(REFRESH_KEY);
  },
};
