import { render, screen } from '@testing-library/react';
import { InterviewCard } from '../InterviewCard';
import type { Interview } from '@/types';

const mockInterview: Interview = {
  id: 'interview-1',
  match_id: 'match-1',
  scheduled_at: '2026-05-01T10:00:00Z',
  format: 'video',
  location: 'https://meet.google.com/abc-defg-hij',
  status: 'scheduled',
};

describe('InterviewCard', () => {
  describe('rendering', () => {
    it('renders interview date and time', () => {
      render(<InterviewCard interview={mockInterview} />);
      expect(screen.getByText('May 1, 2026')).toBeInTheDocument();
      expect(screen.getByText('10:00 AM')).toBeInTheDocument();
    });

    it('renders interview format', () => {
      render(<InterviewCard interview={mockInterview} />);
      expect(screen.getByText('Video')).toBeInTheDocument();
    });

    it('renders interview status', () => {
      render(<InterviewCard interview={mockInterview} />);
      expect(screen.getByText('Scheduled')).toBeInTheDocument();
    });
  });

  describe('location link', () => {
    it('renders join link for video interviews', () => {
      render(<InterviewCard interview={mockInterview} />);
      const link = screen.getByRole('link', { name: /Join/i });
      expect(link).toHaveAttribute('href', 'https://meet.google.com/abc-defg-hij');
    });
  });
});