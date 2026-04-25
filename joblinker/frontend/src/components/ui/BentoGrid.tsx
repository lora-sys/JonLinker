'use client';

import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}

interface BentoGridProps {
  children: React.ReactNode;
  className?: string;
}

export function BentoGrid({ children, className }: BentoGridProps) {
  return (
    <div className={cn(
      "grid grid-cols-1 md:grid-cols-3 gap-4 auto-rows-[minmax(180px,auto)]",
      className
    )}>
      {children}
    </div>
  );
}

interface BentoCardProps {
  children: React.ReactNode;
  className?: string;
  span?: 'default' | 'wide' | 'tall';
}

export function BentoCard({ children, className, span = 'default' }: BentoCardProps) {
  const spanClasses = {
    default: "",
    wide: "md:col-span-2",
    tall: "md:row-span-2",
  };

  return (
    <div className={cn(
      "bg-white/80 backdrop-blur-xl border border-white/20 rounded-2xl p-6 shadow-sm",
      spanClasses[span],
      className
    )}>
      {children}
    </div>
  );
}
