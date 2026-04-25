'use client';

import React, { HTMLAttributes, forwardRef, memo } from 'react';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'outlined' | 'elevated' | 'glass';
  padding?: 'none' | 'sm' | 'md' | 'lg';
  hover?: boolean;
}

const CardComponent = forwardRef<HTMLDivElement, CardProps>(
  (
    { className = '', variant = 'default', padding = 'md', hover = true, children, ...props },
    ref
  ) => {
    const variants = {
      default: 'bg-white/80 backdrop-blur-xl border border-white/20 shadow-lg',
      outlined: 'bg-white border-2 border-slate-200',
      elevated: 'bg-white shadow-xl',
      glass: 'bg-white/80 backdrop-blur-xl border border-white/20',
    };

    const paddings = {
      none: '',
      sm: 'p-3',
      md: 'p-5',
      lg: 'p-6',
    };

    const hoverClasses = hover
      ? 'cursor-pointer transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg active:scale-[0.98]'
      : '';

    return (
      <div
        ref={ref}
        className={`rounded-xl ${variants[variant]} ${paddings[padding]} ${hoverClasses} ${className}`}
        {...props}
      >
        {children}
      </div>
    );
  }
);

CardComponent.displayName = 'Card';

const Card = memo(CardComponent);

export { Card };
export default Card;