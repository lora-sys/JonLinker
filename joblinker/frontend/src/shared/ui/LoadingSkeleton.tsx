import React from 'react'

import Skeleton, { SkeletonCard } from './Skeleton'

interface LoadingSkeletonProps {
  count?: number
  variant?: 'card' | 'list' | 'detail'
}

export default function LoadingSkeleton({ count = 3, variant = 'card' }: LoadingSkeletonProps) {
  if (variant === 'list') {
    return (
      <div className="space-y-3">
        {Array.from({ length: count }).map((_, i) => (
          <SkeletonTableRow key={i} />
        ))}
      </div>
    )
  }

  if (variant === 'detail') {
    return (
      <div className="bg-white rounded-xl border border-slate-200 p-6 space-y-4">
        <Skeleton height={32} width="40%" />
        <Skeleton height={20} width="60%" />
        <div className="flex gap-2">
          <Skeleton width={80} height={32} />
          <Skeleton width={80} height={32} />
          <Skeleton width={80} height={32} />
        </div>
        <Skeleton height={100} />
        <Skeleton height={60} />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonCard key={i} />
      ))}
    </div>
  )
}

function SkeletonTableRow() {
  return (
    <div className="bg-white rounded-lg border border-slate-200 p-4">
      <div className="flex items-center gap-4">
        <Skeleton width={48} height={48} variant="circular" />
        <div className="flex-1 space-y-2">
          <Skeleton height={16} width="30%" />
          <Skeleton height={12} width="20%" />
        </div>
        <Skeleton width={80} height={32} />
      </div>
    </div>
  )
}
