'use client';

import { motion, useInView } from 'framer-motion';
import { useRef } from 'react';

interface HeroHighlightProps {
  children: React.ReactNode;
  className?: string;
  highlightClassName?: string;
}

export function HeroHighlight({
  children,
  className = "",
  highlightClassName = "",
}: HeroHighlightProps) {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true });

  return (
    <motion.span
      ref={ref}
      initial={{ backgroundSize: "0% 100%" }}
      animate={isInView ? { backgroundSize: "100% 100%" } : {}}
      transition={{ duration: 0.8, ease: "easeOut" }}
      className={cn(
        "relative inline-block pb-1 bg-gradient-to-r from-blue-600 to-sky-400 bg-no-repeat [background-position:0_100%] [background-size:100%_40%]",
        className
      )}
    >
      <motion.span
        initial={{ opacity: 0 }}
        animate={isInView ? { opacity: 1 } : {}}
        transition={{ duration: 0.5, delay: 0.2 }}
        className="relative z-10"
      >
        {children}
      </motion.span>
      <motion.span
        className={cn(
          "absolute inset-0 bg-gradient-to-r from-blue-600 to-sky-400 blur-lg opacity-30 -z-10",
          highlightClassName
        )}
        initial={{ opacity: 0 }}
        animate={isInView ? { opacity: 0.3 } : {}}
        transition={{ duration: 0.8, delay: 0.3 }}
      />
    </motion.span>
  );
}

import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}
