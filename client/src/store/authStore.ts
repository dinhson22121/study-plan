import { create } from 'zustand';

import { authApi } from '@/api/auth';
import { setUnauthorizedHandler } from '@/api/client';
import { goalApi } from '@/api/goal';
import { userApi } from '@/api/user';
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  saveTokens,
} from '@/lib/tokenStore';
import type { TokenPairDTO, User } from '@/api/types';

type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated';

type AuthState = {
  status: AuthStatus;
  user: User | null;
  onboarded: boolean;
  bootstrap: () => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  setOnboarded: (value: boolean) => void;
};

async function persist(pair: TokenPairDTO) {
  await saveTokens({ accessToken: pair.access_token, refreshToken: pair.refresh_token });
}

// A user is "onboarded" once they have a study goal set.
async function hasGoal(): Promise<boolean> {
  try {
    const goal = await goalApi.get();
    return Boolean(goal?.goal_type);
  } catch {
    return false;
  }
}

export const useAuthStore = create<AuthState>((set, get) => ({
  status: 'loading',
  user: null,
  onboarded: false,

  bootstrap: async () => {
    const token = await getAccessToken();
    if (!token) {
      set({ status: 'unauthenticated', user: null });
      return;
    }
    try {
      const user = await userApi.me();
      const onboarded = await hasGoal();
      set({ status: 'authenticated', user, onboarded });
    } catch {
      set({ status: 'unauthenticated', user: null });
    }
  },

  login: async (email, password) => {
    const pair = await authApi.login(email, password);
    await persist(pair);
    const user = await userApi.me();
    const onboarded = await hasGoal();
    set({ status: 'authenticated', user, onboarded });
  },

  register: async (email, password) => {
    const pair = await authApi.register(email, password);
    await persist(pair);
    const user = await userApi.me();
    set({ status: 'authenticated', user, onboarded: false });
  },

  logout: async () => {
    const refreshToken = await getRefreshToken();
    if (refreshToken) {
      try {
        await authApi.logout(refreshToken);
      } catch {
        // best-effort server revocation
      }
    }
    await clearTokens();
    set({ status: 'unauthenticated', user: null });
  },

  refreshUser: async () => {
    try {
      const user = await userApi.me();
      set({ user });
    } catch {
      // ignore
    }
  },

  setOnboarded: (value) => set({ onboarded: value }),
}));

// When the API client gives up on refresh, force the app back to logged-out.
setUnauthorizedHandler(() => {
  useAuthStore.setState({ status: 'unauthenticated', user: null });
});
