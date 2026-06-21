import { authApi } from '@/api/auth';
import { userApi } from '@/api/user';
import { goalApi } from '@/api/goal';
import * as tokenStore from '@/lib/tokenStore';
import { useAuthStore } from '@/store/authStore';

jest.mock('@/api/auth');
jest.mock('@/api/user');
jest.mock('@/api/goal');

const mockedAuth = authApi as jest.Mocked<typeof authApi>;
const mockedUser = userApi as jest.Mocked<typeof userApi>;
const mockedGoal = goalApi as jest.Mocked<typeof goalApi>;

const pair = {
  access_token: 'a',
  refresh_token: 'r',
  expires_at: new Date().toISOString(),
};
const user = { id: 'u1', email: 'a@b.com', role: 'STUDENT' as const };

beforeEach(async () => {
  jest.clearAllMocks();
  await tokenStore.clearTokens();
  useAuthStore.setState({ status: 'loading', user: null, onboarded: false });
});

describe('authStore', () => {
  it('login authenticates, persists tokens, and derives onboarded from goal', async () => {
    mockedAuth.login.mockResolvedValue(pair);
    mockedUser.me.mockResolvedValue(user);
    mockedGoal.get.mockResolvedValue({ goal_type: 'MATH_SCORE' });

    await useAuthStore.getState().login('a@b.com', 'password12');

    const s = useAuthStore.getState();
    expect(s.status).toBe('authenticated');
    expect(s.user).toEqual(user);
    expect(s.onboarded).toBe(true);
    expect(await tokenStore.getAccessToken()).toBe('a');
  });

  it('register starts unonboarded (no goal yet)', async () => {
    mockedAuth.register.mockResolvedValue(pair);
    mockedUser.me.mockResolvedValue(user);

    await useAuthStore.getState().register('a@b.com', 'password12');

    expect(useAuthStore.getState().onboarded).toBe(false);
    expect(mockedGoal.get).not.toHaveBeenCalled();
  });

  it('bootstrap with no token is unauthenticated', async () => {
    await useAuthStore.getState().bootstrap();
    expect(useAuthStore.getState().status).toBe('unauthenticated');
    expect(mockedUser.me).not.toHaveBeenCalled();
  });

  it('logout revokes refresh server-side and clears tokens', async () => {
    await tokenStore.saveTokens({ accessToken: 'a', refreshToken: 'r' });
    mockedAuth.logout.mockResolvedValue({ message: 'ok' });
    useAuthStore.setState({ status: 'authenticated', user, onboarded: true });

    await useAuthStore.getState().logout();

    expect(mockedAuth.logout).toHaveBeenCalledWith('r');
    expect(useAuthStore.getState().status).toBe('unauthenticated');
    expect(await tokenStore.getRefreshToken()).toBeNull();
  });
});
