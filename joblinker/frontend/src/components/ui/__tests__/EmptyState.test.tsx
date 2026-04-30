import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import EmptyState from '../EmptyState';
import { Inbox } from 'lucide-react';

describe('EmptyState', () => {
  describe('required props', () => {
    it('renders title', () => {
      render(<EmptyState title="No items found" />);
      expect(screen.getByText('No items found')).toBeInTheDocument();
    });
  });

  describe('optional props', () => {
    it('renders description when provided', () => {
      render(<EmptyState title="No items" description="Try adding some items" />);
      expect(screen.getByText('Try adding some items')).toBeInTheDocument();
    });

    it('renders custom icon when provided', () => {
      render(<EmptyState title="Empty" icon={<Inbox data-testid="custom-icon" />} />);
      expect(screen.getByTestId('custom-icon')).toBeInTheDocument();
    });

    it('renders default icon when none provided', () => {
      const { container } = render(<EmptyState title="Empty" />);
      expect(container.querySelector('svg')).toBeTruthy();
    });
  });

  describe('action buttons', () => {
    it('renders action via href link', () => {
      render(<EmptyState title="Empty" actionLabel="Add Item" href="/items/new" />);
      expect(screen.getByRole('link', { name: /Add Item/i })).toHaveAttribute('href', '/items/new');
    });

    it('renders action via actionHref prop', () => {
      render(<EmptyState title="Empty" actionLabel="Add" actionHref="/add" />);
      expect(screen.getByRole('link', { name: /Add/i })).toHaveAttribute('href', '/add');
    });

    it('renders onAction callback button', async () => {
      const handleAction = jest.fn();
      render(<EmptyState title="Empty" actionLabel="Click me" onAction={handleAction} />);
      await userEvent.click(screen.getByRole('button', { name: /Click me/i }));
      expect(handleAction).toHaveBeenCalledTimes(1);
    });

    it('does not render action when no label provided', () => {
      const { container } = render(<EmptyState title="Empty" />);
      expect(container.querySelector('button')).toBeNull();
      expect(container.querySelector('a')).toBeNull();
    });
  });

  describe('className prop', () => {
    it('applies custom className', () => {
      const { container } = render(<EmptyState title="Empty" className="custom-class" />);
      expect(container.querySelector('.custom-class')).toBeTruthy();
    });
  });

  describe('layout', () => {
    it('centers content', () => {
      const { container } = render(<EmptyState title="Empty" />);
      expect(container.querySelector('.flex-col')).toBeTruthy();
      expect(container.querySelector('.items-center')).toBeTruthy();
      expect(container.querySelector('.justify-center')).toBeTruthy();
    });
  });
});