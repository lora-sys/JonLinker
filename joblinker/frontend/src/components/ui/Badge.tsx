'use client';

import React from 'react';

type BadgeStatus = 'active' | 'paused' | 'pending' | 'negotiating' | 'hired';

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  status?: BadgeStatus;
}

const statusClasses: Record<BadgeStatus, string> = {
  active: 'bg-green-100 text-green-800 border-green-200',
  paused: 'bg-slate-100 text-slate-600 border-slate-200',
  pending: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  negotiating: 'bg-blue-100 text-blue-800 border-blue-200',
  hired: 'bg-emerald-100 text-emerald-800 border-emerald-200',
};

const statusDots: Record<BadgeStatus, string> = {
  active: 'bg-green-500',
  paused: 'bg-slate-400',
  pending: 'bg-yellow-500',
  negotiating: 'bg-blue-500',
  hired: 'bg-emerald-500',
};

export default function Badge({ className = '', status = 'pending', children, ...props }: BadgeProps) {
  return (
    <span
      className={`
        inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border
        ${statusClasses[status]}
        ${className}
      `}
      {...props}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${statusDots[status]}`} />
      {children}
    </span>
  );
}
