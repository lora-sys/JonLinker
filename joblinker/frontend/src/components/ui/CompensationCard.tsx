'use client';

import React from 'react';
import Card from './Card';

interface CompensationBreakdown {
  base: number;
  bonus?: number;
  equity?: number;
  benefits?: number;
  other?: number;
}

interface CompensationCardProps extends React.HTMLAttributes<HTMLDivElement> {
  breakdown: CompensationBreakdown;
  currency?: string;
  period?: 'yearly' | 'monthly';
}

export default function CompensationCard({
  breakdown,
  currency = '$',
  period = 'yearly',
  className = '',
  ...props
}: CompensationCardProps) {
  const formatNumber = (num: number) => {
    return `${currency}${num.toLocaleString()}`;
  };

  const total =
    breakdown.base +
    (breakdown.bonus || 0) +
    (breakdown.equity || 0) +
    (breakdown.benefits || 0) +
    (breakdown.other || 0);

  const items = [
    { label: 'Base Salary', value: breakdown.base, color: 'text-slate-900' },
    breakdown.bonus && { label: 'Signing Bonus', value: breakdown.bonus, color: 'text-green-600' },
    breakdown.equity && { label: 'Equity', value: breakdown.equity, color: 'text-blue-600' },
    breakdown.benefits && { label: 'Benefits', value: breakdown.benefits, color: 'text-purple-600' },
    breakdown.other && { label: 'Other', value: breakdown.other, color: 'text-slate-600' },
  ].filter(Boolean) as { label: string; value: number; color: string }[];

  return (
    <Card className={`p-4 ${className}`} {...props}>
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm text-slate-600">Total Compensation</span>
          <span className="text-lg font-bold text-slate-900">
            {formatNumber(total)}/{period === 'yearly' ? 'yr' : 'mo'}
          </span>
        </div>
        <div className="h-px bg-slate-200" />
        <div className="space-y-2">
          {items.map(({ label, value, color }) => (
            <div key={label} className="flex items-center justify-between text-sm">
              <span className="text-slate-500">{label}</span>
              <span className={`font-medium ${color}`}>{formatNumber(value)}</span>
            </div>
          ))}
        </div>
      </div>
    </Card>
  );
}
