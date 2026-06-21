import { fireEvent, render } from '@testing-library/react-native';

import { Button } from '@/components/ui/Button';

describe('Button', () => {
  it('calls onPress when tapped', () => {
    const onPress = jest.fn();
    const { getByText } = render(<Button title="Đăng nhập" onPress={onPress} />);
    fireEvent.press(getByText('Đăng nhập'));
    expect(onPress).toHaveBeenCalledTimes(1);
  });

  it('does not call onPress when disabled', () => {
    const onPress = jest.fn();
    const { getByText } = render(
      <Button title="Lưu" onPress={onPress} disabled />,
    );
    fireEvent.press(getByText('Lưu'));
    expect(onPress).not.toHaveBeenCalled();
  });

  it('shows a spinner and blocks press while loading', () => {
    const onPress = jest.fn();
    const { queryByText, getByTestId } = render(
      <Button title="Gửi" onPress={onPress} loading testID="submit" />,
    );
    expect(queryByText('Gửi')).toBeNull();
    fireEvent.press(getByTestId('submit'));
    expect(onPress).not.toHaveBeenCalled();
  });
});
