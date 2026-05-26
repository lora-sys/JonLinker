'use client'

import { Plus } from 'lucide-react'
import React from 'react'

interface FABProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: React.ReactNode
  label?: string
  position?: 'bottom-right' | 'bottom-left' | 'bottom-center'
}

const positionClasses = {
  'bottom-right': 'bottom-6 right-6 md:bottom-8 md:right-8',
  'bottom-left': 'bottom-6 left-6 md:bottom-8 md:left-8',
  'bottom-center': 'bottom-6 left-1/2 -translate-x-1/2',
}

export default function FAB({
  icon,
  label,
  position = 'bottom-right',
  className = '',
  ...props
}: FABProps) {
  return (
    <button
      className={`
        fixed ${positionClasses[position]}
        w-14 h-14 rounded-full
        bg-blue-600 text-white shadow-lg
        flex items-center justify-center
        hover:bg-blue-700 hover:shadow-xl
        active:scale-95
        transition-all duration-200
        focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
        cursor-pointer
        ${className}
      `}
      aria-label={label || 'Floating action button'}
      {...props}
    >
      {icon || <Plus className="w-6 h-6" />}
    </button>
  )
}
