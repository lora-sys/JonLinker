'use client';

import React from 'react';

interface SkillTagProps extends React.HTMLAttributes<HTMLSpanElement> {
  color?: 'blue' | 'green' | 'purple' | 'orange' | 'pink' | 'slate';
  size?: 'sm' | 'md';
}

const colorClasses = {
  blue: 'bg-blue-100 text-blue-700 border-blue-200',
  green: 'bg-green-100 text-green-700 border-green-200',
  purple: 'bg-purple-100 text-purple-700 border-purple-200',
  orange: 'bg-orange-100 text-orange-700 border-orange-200',
  pink: 'bg-pink-100 text-pink-700 border-pink-200',
  slate: 'bg-slate-100 text-slate-700 border-slate-200',
};

const sizeClasses = {
  sm: 'px-2 py-0.5 text-xs',
  md: 'px-2.5 py-1 text-xs',
};

// Predefined skill color mappings for common skills
const skillColorMap: Record<string, SkillTagProps['color']> = {
  react: 'blue',
  typescript: 'blue',
  javascript: 'blue',
  node: 'green',
  python: 'green',
  go: 'green',
  rust: 'orange',
  sql: 'slate',
  aws: 'orange',
  docker: 'blue',
  kubernetes: 'purple',
  graphql: 'pink',
  vue: 'green',
  angular: 'pink',
  java: 'orange',
  csharp: 'purple',
  ruby: 'pink',
  php: 'purple',
  swift: 'orange',
  kotlin: 'purple',
};

export default function SkillTag({ color, size = 'sm', children, className = '', ...props }: SkillTagProps) {
  // Auto-assign color based on skill name if not specified
  const skillName = typeof children === 'string' ? children.toLowerCase() : '';
  const resolvedColor = color || skillColorMap[skillName] || 'slate';

  return (
    <span
      className={`
        inline-flex items-center rounded-full border font-medium
        ${colorClasses[resolvedColor]}
        ${sizeClasses[size]}
        ${className}
      `}
      {...props}
    >
      {children}
    </span>
  );
}
