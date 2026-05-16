import { render, screen } from '@testing-library/react'

import Avatar from '../Avatar'

describe('avatar', () => {
  describe('type prop', () => {
    it('renders person type by default', () => {
      const { container } = render(<Avatar />)
      expect(container.firstChild).toHaveClass('bg-slate-100', 'text-slate-600')
    })

    it('renders bot type', () => {
      const { container } = render(<Avatar type="bot" />)
      expect(container.firstChild).toHaveClass('bg-blue-100', 'text-blue-600')
    })

    it('renders seeker type', () => {
      const { container } = render(<Avatar type="seeker" />)
      expect(container.firstChild).toHaveClass('bg-emerald-100', 'text-emerald-600')
    })

    it('renders recruiter type', () => {
      const { container } = render(<Avatar type="recruiter" />)
      expect(container.firstChild).toHaveClass('bg-purple-100', 'text-purple-600')
    })
  })

  describe('size prop', () => {
    it('renders sm size', () => {
      const { container } = render(<Avatar size="sm" />)
      expect(container.firstChild).toHaveClass('w-8', 'h-8', 'text-xs')
    })

    it('renders md size', () => {
      const { container } = render(<Avatar size="md" />)
      expect(container.firstChild).toHaveClass('w-10', 'h-10', 'text-sm')
    })

    it('renders lg size', () => {
      const { container } = render(<Avatar size="lg" />)
      expect(container.firstChild).toHaveClass('w-12', 'h-12', 'text-base')
    })

    it('renders xl size', () => {
      const { container } = render(<Avatar size="xl" />)
      expect(container.firstChild).toHaveClass('w-16', 'h-16', 'text-lg')
    })
  })

  describe('initials prop', () => {
    it('displays initials when provided', () => {
      render(<Avatar initials="JD" />)
      expect(screen.getByText('JD')).toBeInTheDocument()
    })

    it('shows type icon when no initials', () => {
      const { container } = render(<Avatar type="bot" />)
      expect(container.querySelector('svg')).toBeTruthy()
    })
  })

  describe('className prop', () => {
    it('applies custom className', () => {
      const { container } = render(<Avatar className="custom-class" />)
      expect(container.firstChild).toHaveClass('custom-class')
    })
  })
})
