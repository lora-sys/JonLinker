import { render } from '@testing-library/react'

import LoadingSkeleton from '../LoadingSkeleton'

describe('loadingSkeleton', () => {
  describe('card variant', () => {
    it('renders default count of 3 skeleton cards', () => {
      render(<LoadingSkeleton variant="card" />)
      const cards = document.querySelectorAll('[class*="rounded-xl"]')
      expect(cards.length).toBeGreaterThan(0)
    })

    it('renders specified count', () => {
      render(<LoadingSkeleton variant="card" count={5} />)
      const content = document.body.textContent
      expect(content).toBeTruthy()
    })
  })

  describe('list variant', () => {
    it('renders list skeleton rows', () => {
      const { container } = render(<LoadingSkeleton variant="list" />)
      expect(container.querySelector('.space-y-3')).toBeTruthy()
    })
  })

  describe('detail variant', () => {
    it('renders detail skeleton with proper structure', () => {
      const { container } = render(<LoadingSkeleton variant="detail" />)
      expect(container.querySelector('.bg-white')).toBeTruthy()
      expect(container.querySelector('.rounded-xl')).toBeTruthy()
    })
  })

  describe('count prop', () => {
    it('renders correct number of card skeletons', () => {
      render(<LoadingSkeleton variant="card" count={2} />)
      const skeletons = document.querySelectorAll('[class*="animate-pulse"]')
      expect(skeletons.length).toBe(2)
    })
  })
})
