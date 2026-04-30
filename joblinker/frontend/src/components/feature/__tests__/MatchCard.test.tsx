import { render, screen } from '@testing-library/react';
import MatchCard from '../MatchCard';
import type { Match } from '@/types';

const mockMatch: Match = {
  id: 'match-1',
  seeker_agent_id: 'seeker-1',
  job_id: 'job-1',
  score: 0.92,
  status: 'mutual_interest',
  created_at: '2026-04-28T00:00:00Z',
  updated_at: '2026-04-28T00:00:00Z',
};

describe('MatchCard', () => {
  describe('rendering', () => {
    it('renders match score as percentage', () => {
      render(<MatchCard match={mockMatch} />);
      expect(screen.getByText('92%')).toBeInTheDocument();
    });

    it('renders match status', () => {
      render(<MatchCard match={mockMatch} />);
      expect(screen.getByText('Mutual Interest')).toBeInTheDocument();
    });
  });

  describe('score badge', () => {
    it('displays high score in green', () => {
      const { container } = render(<MatchCard match={mockMatch} />);
      expect(container.querySelector('.bg-emerald-100')).toBeTruthy();
    });

    it('displays medium score in yellow', () => {
      const mediumMatch = { ...mockMatch, score: 0.65 };
      const { container } = render(<MatchCard match={mediumMatch} />);
      expect(container.querySelector('.bg-yellow-100')).toBeTruthy();
    });

    it('displays low score in red', () => {
      const lowMatch = { ...mockMatch, score: 0.35 };
      const { container } = render(<MatchCard match={lowMatch} />);
      expect(container.querySelector('.bg-red-100')).toBeTruthy();
    });
  });
});