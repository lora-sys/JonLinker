'use client'

import React from 'react'

interface ScoreBarProps extends React.HTMLAttributes<HTMLDivElement> {
  score: number // 0-100
  showLabel?: boolean
  size?: 'sm' | 'md' | 'lg'
}

const sizeClasses = {
  sm: 'h-1.5',
  md: 'h-2.5',
  lg: 'h-4',
}

export default function ScoreBar({ score, showLabel = true, size = 'md', className = '', ...props }: ScoreBarProps) {
  // Clamp score between 0 and 100
  const clampedScore = Math.min(100, Math.max(0, score))

  // Color thresholds: >80 green, 50-80 yellow, <50 red
  const getColorClass = (s: number) => {
    if (s >= 80)
      return 'bg-green-500'
    if (s >= 50)
      return 'bg-yellow-500'
    return 'bg-red-500'
  }

  return (
    <div className={`flex items-center gap-2 ${className}`} {...props}>
      <div className={`flex-1 bg-slate-200 rounded-full overflow-hidden ${sizeClasses[size]}`}>
        <div
          className={`h-full rounded-full transition-all duration-500 ${getColorClass(clampedScore)}`}
          style={{ width: `${clampedScore}%` }}
        />
      </div>
      {showLabel && (
        <span className="text-xs font-medium text-slate-600 min-w-[2.5rem] text-right">
          {clampedScore}
          %
        </span>
      )}
    </div>
  )
}
