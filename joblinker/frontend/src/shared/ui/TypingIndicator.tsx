'use client'

import React from 'react'

interface TypingIndicatorProps extends React.HTMLAttributes<HTMLDivElement> {
  name?: string
}

export default function TypingIndicator({ name, className = '', ...props }: TypingIndicatorProps) {
  return (
    <div className={`flex items-center gap-2 ${className}`} {...props}>
      <div className="flex gap-1">
        {[0, 1, 2].map(i => (
          <span
            key={i}
            className="w-2 h-2 bg-slate-400 rounded-full animate-bounce"
            style={{ animationDelay: `${i * 150}ms` }}
          />
        ))}
      </div>
      {name && (
        <span className="text-xs text-slate-500">
          {name}
          {' '}
          is typing...
        </span>
      )}
    </div>
  )
}
