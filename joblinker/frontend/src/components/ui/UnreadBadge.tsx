'use client';

import React from 'react';

interface UnreadBadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  count?: number;
  max?: number;
}

export default function UnreadBadge({ count = 0, max = 99, className = '', ...props }: UnreadBadgeProps) {
  if (count === 0) return null;

  const displayCount = count > max ? `${max}+` : count;

  return (
    <span
      className={`
        inline-flex items-center justify-center
        min-w-[1.25rem] h-5 px-1.5
        text-xs font-bold text-white
        bg-red-500 rounded-full
        ${className}
      `}
      {...props}
    >
      {displayCount}
    </span>
  );
}
