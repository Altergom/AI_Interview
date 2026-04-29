import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Input } from '../components/common/Input';

describe('Input Component', () => {
  it('应该渲染输入框', () => {
    render(<Input placeholder="Enter text" />);
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument();
  });

  it('应该显示标签', () => {
    render(<Input label="Username" />);
    expect(screen.getByText('Username')).toBeInTheDocument();
  });

  it('应该显示错误信息', () => {
    render(<Input error="This field is required" />);
    expect(screen.getByText('This field is required')).toBeInTheDocument();
  });

  it('应该显示辅助文本', () => {
    render(<Input helperText="Enter your email address" />);
    expect(screen.getByText('Enter your email address')).toBeInTheDocument();
  });

  it('应该响应输入变化', async () => {
    const handleChange = vi.fn();
    const user = userEvent.setup();

    render(<Input onChange={handleChange} />);

    const input = screen.getByRole('textbox');
    await user.type(input, 'test');

    expect(handleChange).toHaveBeenCalled();
  });

  it('应该支持密码显示/隐藏切换', async () => {
    const user = userEvent.setup();

    render(<Input type="password" value="password123" readOnly />);

    const input = screen.getByDisplayValue('password123');
    expect(input).toHaveAttribute('type', 'password');

    const toggleButton = screen.getByRole('button');
    await user.click(toggleButton);

    expect(input).toHaveAttribute('type', 'text');
  });

  it('应该在禁用时不可编辑', () => {
    render(<Input disabled />);
    const input = screen.getByRole('textbox');
    expect(input).toBeDisabled();
  });
});
