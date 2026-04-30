import { render, screen } from '@testing-library/react';
import MessageBubble from '../MessageBubble';

describe('MessageBubble', () => {
  describe('content prop', () => {
    it('renders message content', () => {
      render(<MessageBubble content="Hello world" sender="user" />);
      expect(screen.getByText('Hello world')).toBeInTheDocument();
    });
  });

  describe('sender prop', () => {
    it('aligns user messages to the right', () => {
      const { container } = render(<MessageBubble content="Hi" sender="user" />);
      expect(container.querySelector('.flex')).toHaveClass('justify-end');
    });

    it('aligns agent messages to the left', () => {
      const { container } = render(<MessageBubble content="Hi" sender="agent" />);
      expect(container.querySelector('.flex')).toHaveClass('justify-start');
    });

    it('aligns system messages to the left', () => {
      const { container } = render(<MessageBubble content="System msg" sender="system" />);
      expect(container.querySelector('.flex')).toHaveClass('justify-start');
    });
  });

  describe('sender prop with colors', () => {
    it('applies blue background for user sender', () => {
      const { container } = render(<MessageBubble content="Hi" sender="user" />);
      expect(container.querySelector('.bg-blue-100')).toBeTruthy();
    });

    it('applies gray background for agent sender', () => {
      const { container } = render(<MessageBubble content="Hi" sender="agent" />);
      expect(container.querySelector('.bg-gray-50')).toBeTruthy();
    });
  });

  describe('intent prop', () => {
    it('renders intent tag when provided', () => {
      render(<MessageBubble content="Offer" sender="agent" intent="OFFER" />);
      expect(screen.getByText('OFFER')).toBeInTheDocument();
    });

    it('applies correct color for OFFER intent', () => {
      const { container } = render(<MessageBubble content="Offer" sender="agent" intent="OFFER" />);
      expect(container.querySelector('.bg-purple-50')).toBeTruthy();
    });

    it('applies correct color for NEGOTIATION intent', () => {
      const { container } = render(<MessageBubble content="Let's negotiate" sender="agent" intent="NEGOTIATION" />);
      expect(container.querySelector('.bg-yellow-50')).toBeTruthy();
    });
  });

  describe('timestamp prop', () => {
    it('renders formatted timestamp', () => {
      const date = new Date('2026-04-28T14:30:00');
      render(<MessageBubble content="Hi" sender="user" timestamp={date} />);
      expect(screen.getByText('2:30 PM')).toBeInTheDocument();
    });

    it('does not render timestamp when not provided', () => {
      const { container } = render(<MessageBubble content="Hi" sender="user" />);
      expect(container.querySelector('.whitespace-nowrap')).toBeNull();
    });
  });

  describe('senderName prop', () => {
    it('renders sender name above message', () => {
      render(<MessageBubble content="Hi" sender="agent" senderName="Agent Bob" />);
      expect(screen.getByText('Agent Bob')).toBeInTheDocument();
    });
  });
});