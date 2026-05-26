import { render, screen } from '@testing-library/react'

import { Card } from '../Card'

describe('card', () => {
  describe('rendering', () => {
    it('renders with children', () => {
      render(<Card>Test Content</Card>)
      expect(screen.getByText('Test Content')).toBeInTheDocument()
    })

    it('renders with default variant', () => {
      const { container } = render(<Card>Content</Card>)
      expect(container.firstChild).toBeTruthy()
    })

    it('renders with className override', () => {
      const { container } = render(<Card className="custom-class">Content</Card>)
      expect(container.firstChild).toHaveClass('custom-class')
    })
  })

  describe('variant prop', () => {
    it('applies correct classes for default variant', () => {
      const { container } = render(<Card variant="default">Default</Card>)
      expect(container.firstChild).toHaveClass('bg-white/80')
    })

    it('applies correct classes for outlined variant', () => {
      const { container } = render(<Card variant="outlined">Outlined</Card>)
      expect(container.firstChild).toHaveClass('bg-white', 'border-2', 'border-slate-200')
    })

    it('applies correct classes for elevated variant', () => {
      const { container } = render(<Card variant="elevated">Elevated</Card>)
      expect(container.firstChild).toHaveClass('bg-white', 'shadow-xl')
    })

    it('applies correct classes for glass variant', () => {
      const { container } = render(<Card variant="glass">Glass</Card>)
      expect(container.firstChild).toHaveClass('backdrop-blur-xl')
    })
  })

  describe('padding prop', () => {
    it('applies no padding when padding="none"', () => {
      const { container } = render(<Card padding="none">NoPad</Card>)
      expect(container.firstChild).toHaveClass('p-0')
    })

    it('applies sm padding', () => {
      const { container } = render(<Card padding="sm">SmPad</Card>)
      expect(container.firstChild).toHaveClass('p-3')
    })

    it('applies md padding (default)', () => {
      const { container } = render(<Card padding="md">MdPad</Card>)
      expect(container.firstChild).toHaveClass('p-5')
    })

    it('applies lg padding', () => {
      const { container } = render(<Card padding="lg">LgPad</Card>)
      expect(container.firstChild).toHaveClass('p-6')
    })
  })

  describe('hover prop', () => {
    it('applies hover classes when hover=true', () => {
      const { container } = render(<Card hover={true}>Hover</Card>)
      expect(container.firstChild).toHaveClass('hover:-translate-y-0.5', 'hover:shadow-lg')
    })

    it('does not apply hover classes when hover=false', () => {
      const { container } = render(<Card hover={false}>NoHover</Card>)
      expect(container.firstChild).not.toHaveClass('hover:-translate-y-0.5')
    })
  })

  describe('forwardRef', () => {
    it('forwards ref correctly', () => {
      const ref = { current: null } as React.RefObject<HTMLDivElement>
      render(<Card ref={ref}>RefTest</Card>)
      expect(ref.current).toBeTruthy()
    })
  })
})
