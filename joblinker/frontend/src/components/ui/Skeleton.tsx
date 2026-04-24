'use client';

import React from 'react';

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  width?: string | number;
  height?: string | number;
  variant?: 'text' | 'circular' | 'rectangular';
}

export default function Skeleton({
  className = '',
  width,
  height,
  variant = 'rectangular',
  style,
  ...props
}: SkeletonProps) {
  const variantClasses = {
    text: 'rounded',
    circular: 'rounded-full',
    rectangular: 'rounded-lg',
  };

  return (
    <div
      className={`
        animate-pulse bg-slate-200
        ${variantClasses[variant]}
        ${className}
      `}
      style={{
        width,
        height,
        ...style,
      }}
      {...props}
    />
  );
}

// Preset skeleton patterns
export function SkeletonCard() {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4 space-y-3">
      <Skeleton height={20} width="60%" />
      <Skeleton height={16} width="40%" />
      <div className="flex gap-2">
        <Skeleton width={60} height={24} variant="text" />
        <Skeleton width={80} height={24} variant="text" />
      </div>
    </div>
  );
}

export function SkeletonTableRow() {
  return (
    <div className="flex items-center gap-4 p-4 border-b border-slate-100">
      <Skeleton width={40} height={40} variant="circular" />
      <div className="flex-1 space-y-2">
        <Skeleton height={16} width="30%" />
        <Skeleton height={12} width="20%" />
      </div>
      <Skeleton width={80} height={24} />
    </div>
  );
}
