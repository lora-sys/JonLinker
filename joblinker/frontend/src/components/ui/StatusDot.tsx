'use client';

import React from 'react';

type DotStatus = 'active' | 'idle' | 'offline' | 'error';

interface StatusDotProps extends React.HTMLAttributes<HTMLSpanElement> {
  status?: DotStatus;
  pulse?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

const statusColors: Record<DotStatus, string> = {
  active: 'bg-green-500',
  idle: 'bg-yellow-500',
  offline: 'bg-slate-400',
  error: 'bg-red-500',
};

const sizeClasses = {
  sm: 'w-2 h-2',
  md: 'w-2.5 h-2.5',
  lg: 'w-3 h-3',
};

export default function StatusDot({ status = 'offline', pulse = true, size = 'md', className = '', ...props }: StatusDotProps) {
  const shouldPulse = pulse && status === 'active';

  return (
    <span className={`relative inline-flex ${className}`} {...props}>
      <span
        className={`
          inline-block rounded-full
          ${statusColors[status]}
          ${sizeClasses[size]}
        `}
      />
      {shouldPulse && (
        <span
          className={`
            absolute inline-block rounded-full
            ${statusColors[status]}
            ${sizeClasses[size]}
            animate-ping
          `}
        />
      )}
    </span>
  );
}
