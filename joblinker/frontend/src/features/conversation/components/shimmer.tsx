'use client'

import type { MotionProps } from 'motion/react'
import type { CSSProperties, ElementType } from 'react'

import { motion } from 'motion/react'
import { memo } from 'react'

import { cn } from '@/lib/utils'

type MotionHTMLProps = MotionProps & Record<string, unknown>

// Pre-defined motion components to satisfy react/static-components
const motionComponents = {
  p: motion.p,
  span: motion.span,
  div: motion.div,
  h1: motion.h1,
  h2: motion.h2,
  h3: motion.h3,
  h4: motion.h4,
  h5: motion.h5,
  h6: motion.h6,
} as const satisfies Record<string, React.ComponentType<MotionHTMLProps>>

type MotionElement = keyof typeof motionComponents

export interface TextShimmerProps {
  children: string
  as?: ElementType
  className?: string
  duration?: number
  spread?: number
}

function ShimmerComponent({
  children,
  as = 'p',
  className,
  duration = 2,
  spread = 2,
}: TextShimmerProps) {
  const MotionComponent = motionComponents[as as MotionElement] ?? motionComponents.p

  const dynamicSpread = (children?.length ?? 0) * spread

  return (
    <MotionComponent
      animate={{ backgroundPosition: '0% center' }}
      className={cn(
        'relative inline-block bg-[length:250%_100%,auto] bg-clip-text text-transparent',
        '[--bg:linear-gradient(90deg,#0000_calc(50%-var(--spread)),var(--color-background),#0000_calc(50%+var(--spread)))] [background-repeat:no-repeat,padding-box]',
        className,
      )}
      initial={{ backgroundPosition: '100% center' }}
      style={
        {
          '--spread': `${dynamicSpread}px`,
          'backgroundImage':
            'var(--bg), linear-gradient(var(--color-muted-foreground), var(--color-muted-foreground))',
        } as CSSProperties
      }
      transition={{
        duration,
        ease: 'linear',
        repeat: Number.POSITIVE_INFINITY,
      }}
    >
      {children}
    </MotionComponent>
  )
}

export const Shimmer = memo(ShimmerComponent)
