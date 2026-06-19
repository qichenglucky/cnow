import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import StatusBadge from '../StatusBadge';

describe('StatusBadge', () => {
  it('renders ready status with success (green) tag', () => {
    render(<StatusBadge status="ready" />);
    const tag = screen.getByText('就绪');
    expect(tag).toBeInTheDocument();
    // antd Tag with color="success" gets class ant-tag-success
    expect(tag.closest('.ant-tag')).toHaveClass('ant-tag-success');
  });

  it('renders deploying status with processing tag', () => {
    render(<StatusBadge status="deploying" />);
    const tag = screen.getByText('部署中');
    expect(tag).toBeInTheDocument();
    expect(tag.closest('.ant-tag')).toHaveClass('ant-tag-processing');
  });

  it('renders failed status with error (red) tag', () => {
    render(<StatusBadge status="failed" />);
    const tag = screen.getByText('失败');
    expect(tag).toBeInTheDocument();
    expect(tag.closest('.ant-tag')).toHaveClass('ant-tag-error');
  });

  it('renders creating status with processing (blue) tag', () => {
    render(<StatusBadge status="creating" />);
    const tag = screen.getByText('创建中');
    expect(tag).toBeInTheDocument();
    expect(tag.closest('.ant-tag')).toHaveClass('ant-tag-processing');
  });

  it('renders unknown status with default tag showing raw status', () => {
    render(<StatusBadge status="foobar" />);
    const tag = screen.getByText('foobar');
    expect(tag).toBeInTheDocument();
    expect(tag.closest('.ant-tag')).toHaveClass('ant-tag-default');
  });
});
