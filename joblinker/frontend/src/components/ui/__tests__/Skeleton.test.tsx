import { render, screen } from '@testing-library/react';
import Skeleton, { SkeletonCard, SkeletonTableRow } from '../Skeleton';

describe('Skeleton', () => {
  describe('variant prop', () => {
    it('renders text variant with rounded class', () => {
      const { container } = render(<Skeleton variant="text" />);
      expect(container.firstChild).toHaveClass('rounded');
    });

    it('renders circular variant with rounded-full', () => {
      const { container } = render(<Skeleton variant="circular" />);
      expect(container.firstChild).toHaveClass('rounded-full');
    });

    it('renders rectangular variant with rounded-lg', () => {
      const { container } = render(<Skeleton variant="rectangular" />);
      expect(container.firstChild).toHaveClass('rounded-lg');
    });
  });

  describe('width and height props', () => {
    it('applies width style', () => {
      const { container } = render(<Skeleton width={100} />);
      expect(container.firstChild).toHaveStyle({ width: 100 });
    });

    it('applies height style', () => {
      const { container } = render(<Skeleton height={40} />);
      expect(container.firstChild).toHaveStyle({ height: 40 });
    });

    it('applies string width', () => {
      const { container } = render(<Skeleton width="60%" />);
      expect(container.firstChild).toHaveStyle({ width: '60%' });
    });
  });

  describe('className prop', () => {
    it('applies custom className', () => {
      const { container } = render(<Skeleton className="custom-skeleton" />);
      expect(container.firstChild).toHaveClass('custom-skeleton');
    });
  });

  describe('SkeletonCard preset', () => {
    it('renders card structure', () => {
      const { container } = render(<SkeletonCard />);
      expect(container.querySelector('.bg-white')).toBeTruthy();
      expect(container.querySelector('.rounded-xl')).toBeTruthy();
    });
  });

  describe('SkeletonTableRow preset', () => {
    it('renders table row structure', () => {
      const { container } = render(<SkeletonTableRow />);
      expect(container.querySelector('.flex')).toBeTruthy();
      expect(container.querySelector('.items-center')).toBeTruthy();
    });

    it('includes circular skeleton for avatar', () => {
      const { container } = render(<SkeletonTableRow />);
      expect(container.querySelector('.rounded-full')).toBeTruthy();
    });
  });
});