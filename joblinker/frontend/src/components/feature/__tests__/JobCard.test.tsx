import { render, screen } from '@testing-library/react'

import type { Job } from '@/types'

import JobCard from '../JobCard'

const mockJob: Job = {
  id: 'job-1',
  agent_id: 'rec-1',
  structured: {
    title: 'Senior React Developer',
    description: 'We are looking for a senior React developer to join our team.',
    requirements: ['React', 'TypeScript', 'Node.js'],
    location: 'San Francisco, CA',
    work_type: 'remote',
    salary_range: { min: 150000, max: 200000, currency: 'USD' },
  },
  status: 'active',
  created_at: '2026-04-28T00:00:00Z',
  updated_at: '2026-04-28T00:00:00Z',
}

describe('jobCard', () => {
  describe('rendering', () => {
    it('renders job title', () => {
      render(<JobCard job={mockJob} />)
      expect(screen.getByText('Senior React Developer')).toBeInTheDocument()
    })

    it('renders job description', () => {
      render(<JobCard job={mockJob} />)
      expect(screen.getByText(/We are looking for a senior React developer/i)).toBeInTheDocument()
    })

    it('renders job location', () => {
      render(<JobCard job={mockJob} />)
      expect(screen.getByText('San Francisco, CA')).toBeInTheDocument()
    })

    it('renders salary range', () => {
      render(<JobCard job={mockJob} />)
      expect(screen.getByText('$150k - $200k')).toBeInTheDocument()
    })
  })

  describe('work_type display', () => {
    it('displays Remote badge for remote jobs', () => {
      const { container } = render(<JobCard job={mockJob} />)
      expect(container.textContent).toContain('Remote')
    })
  })

  describe('requirements display', () => {
    it('renders skill badges for requirements', () => {
      render(<JobCard job={mockJob} />)
      expect(screen.getByText('React')).toBeInTheDocument()
      expect(screen.getByText('TypeScript')).toBeInTheDocument()
    })
  })
})
