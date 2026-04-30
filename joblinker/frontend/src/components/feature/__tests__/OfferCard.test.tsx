import { render, screen } from '@testing-library/react';
import OfferCard from '../OfferCard';
import type { Offer } from '@/types';

const mockOffer: Offer = {
  id: 'offer-1',
  match_id: 'match-1',
  compensation: {
    base_salary: 180000,
    signing_bonus: 15000,
    annual_bonus: 20000,
    equity_grant: 50000,
    currency: 'USD',
    total_compensation: 265000,
  },
  start_date: '2026-06-01',
  status: 'pending',
  created_at: '2026-04-28T00:00:00Z',
  updated_at: '2026-04-28T00:00:00Z',
};

describe('OfferCard', () => {
  describe('rendering', () => {
    it('renders base salary', () => {
      render(<OfferCard offer={mockOffer} />);
      expect(screen.getByText('$180,000')).toBeInTheDocument();
    });

    it('renders start date', () => {
      render(<OfferCard offer={mockOffer} />);
      expect(screen.getByText('2026-06-01')).toBeInTheDocument();
    });

    it('renders offer status', () => {
      render(<OfferCard offer={mockOffer} />);
      expect(screen.getByText('Pending')).toBeInTheDocument();
    });
  });

  describe('compensation breakdown', () => {
    it('displays signing bonus when present', () => {
      render(<OfferCard offer={mockOffer} />);
      expect(screen.getByText('$15,000')).toBeInTheDocument();
    });

    it('displays annual bonus when present', () => {
      render(<OfferCard offer={mockOffer} />);
      expect(screen.getByText('$20,000')).toBeInTheDocument();
    });

    it('displays equity when present', () => {
      render(<OfferCard offer={mockOffer} />);
      expect(screen.getByText('$50,000')).toBeInTheDocument();
    });
  });

  describe('total compensation', () => {
    it('shows total compensation', () => {
      render(<OfferCard offer={mockOffer} />);
      expect(screen.getByText('$265,000')).toBeInTheDocument();
    });
  });
});