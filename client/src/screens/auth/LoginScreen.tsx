// LoginScreen — email/password sign-in via the auth store.
import { useState } from 'react';
import { TextInput, View } from 'react-native';
import { useMutation } from '@tanstack/react-query';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { AppText, Button, Screen } from '@/components/ui';
import { ApiError } from '@/api/client';
import { useAuthStore } from '@/store/authStore';
import { colors, fontFamily, fontSize, radius, spacing } from '@/theme/tokens';
import type { AuthStackParamList } from '@/navigation/types';

type Props = NativeStackScreenProps<AuthStackParamList, 'Login'>;

type Credentials = { email: string; password: string };

function messageForError(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 401) {
      return 'Email hoặc mật khẩu không đúng';
    }
    if (error.status === 0) {
      return 'Không kết nối được máy chủ. Vui lòng thử lại.';
    }
    return error.message;
  }
  return 'Đã có lỗi xảy ra. Vui lòng thử lại.';
}

export default function LoginScreen(_props: Props) {
  const login = useAuthStore((s) => s.login);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  // On success the RootNavigator switches stacks automatically.
  const mutation = useMutation<void, unknown, Credentials>({
    mutationFn: ({ email: e, password: p }) => login(e.trim(), p),
  });

  const canSubmit =
    email.trim().length > 0 && password.length > 0 && !mutation.isPending;

  const handleSubmit = () => {
    if (!canSubmit) {
      return;
    }
    mutation.mutate({ email, password });
  };

  const inputStyle = {
    minHeight: 52,
    borderRadius: radius.button,
    borderWidth: 1,
    borderColor: colors.borderSoft,
    backgroundColor: colors.surfaceCard,
    paddingHorizontal: spacing.lg,
    fontFamily: fontFamily.ui,
    fontSize: fontSize.bodyLarge,
    color: colors.textPrimary,
  } as const;

  return (
    <Screen scroll>
      <View style={{ gap: spacing.xs, marginBottom: spacing.xxl }}>
        <AppText variant="hero" weight="bold" display>
          Đăng nhập
        </AppText>
        <AppText variant="body" color={colors.textSecondary}>
          Tiếp tục lộ trình học cá nhân hóa của bạn.
        </AppText>
      </View>

      <View style={{ gap: spacing.md }}>
        <View style={{ gap: spacing.xs }}>
          <AppText variant="caption" weight="medium" color={colors.textSecondary}>
            Email
          </AppText>
          <TextInput
            value={email}
            onChangeText={setEmail}
            placeholder="ban@example.com"
            placeholderTextColor={colors.textMuted}
            autoCapitalize="none"
            autoCorrect={false}
            keyboardType="email-address"
            textContentType="emailAddress"
            editable={!mutation.isPending}
            style={inputStyle}
            accessibilityLabel="Email"
          />
        </View>

        <View style={{ gap: spacing.xs }}>
          <AppText variant="caption" weight="medium" color={colors.textSecondary}>
            Mật khẩu
          </AppText>
          <TextInput
            value={password}
            onChangeText={setPassword}
            placeholder="••••••••••"
            placeholderTextColor={colors.textMuted}
            secureTextEntry
            autoCapitalize="none"
            autoCorrect={false}
            textContentType="password"
            editable={!mutation.isPending}
            onSubmitEditing={handleSubmit}
            style={inputStyle}
            accessibilityLabel="Mật khẩu"
          />
          <AppText variant="caption" color={colors.textMuted}>
            Mật khẩu gồm 10–72 ký tự, có cả chữ và số.
          </AppText>
        </View>

        {mutation.isError ? (
          <AppText variant="body" weight="medium" color={colors.brandDanger}>
            {messageForError(mutation.error)}
          </AppText>
        ) : null}

        <Button
          title="Đăng nhập"
          onPress={handleSubmit}
          loading={mutation.isPending}
          disabled={!canSubmit}
          style={{ marginTop: spacing.sm }}
          testID="login-submit"
        />
      </View>
    </Screen>
  );
}
