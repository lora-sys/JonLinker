'use client';

import React, { useState } from 'react';

interface NegotiateSliderProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  min?: number;
  max?: number;
  step?: number;
  value?: number;
  onChange?: (value: number) => void;
  showValue?: boolean;
  formatValue?: (value: number) => string;
}

export default function NegotiateSlider({
  min = 0,
  max = 100,
  step = 1,
  value: controlledValue,
  onChange,
  showValue = true,
  formatValue = (v) => `$${v.toLocaleString()}`,
  className = '',
  ...props
}: NegotiateSliderProps) {
  const [internalValue, setInternalValue] = useState((min + max) / 2);
  const value = controlledValue !== undefined ? controlledValue : internalValue;

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = Number(e.target.value);
    setInternalValue(newValue);
    onChange?.(newValue);
  };

  const percentage = ((value - min) / (max - min)) * 100;

  return (
    <div className={`space-y-2 ${className}`}>
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium text-slate-700">Counter Offer</label>
        {showValue && (
          <span className="text-sm font-bold text-blue-600">{formatValue(value)}</span>
        )}
      </div>
      <div className="relative">
        <input
          type="range"
          min={min}
          max={max}
          step={step}
          value={value}
          onChange={handleChange}
          className="w-full h-2 bg-slate-200 rounded-lg appearance-none cursor-pointer accent-blue-600"
          {...props}
        />
        <div
          className="absolute top-0 left-0 h-2 bg-blue-600 rounded-l-lg pointer-events-none"
          style={{ width: `${percentage}%` }}
        />
      </div>
      <div className="flex justify-between text-xs text-slate-500">
        <span>{formatValue(min)}</span>
        <span>{formatValue(max)}</span>
      </div>
    </div>
  );
}
