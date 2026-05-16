'use client'

import React, { useEffect, useState } from 'react'

interface CountdownProps extends React.HTMLAttributes<HTMLDivElement> {
  targetDate: Date | string
  onExpire?: () => void
  showLabels?: boolean
}

function calculateTimeLeft(targetDate: Date): {
  days: number
  hours: number
  minutes: number
  seconds: number
  total: number
} {
  const total = targetDate.getTime() - Date.now()

  if (total <= 0) {
    return { days: 0, hours: 0, minutes: 0, seconds: 0, total: 0 }
  }

  return {
    days: Math.floor(total / (1000 * 60 * 60 * 24)),
    hours: Math.floor((total / (1000 * 60 * 60)) % 24),
    minutes: Math.floor((total / 1000 / 60) % 60),
    seconds: Math.floor((total / 1000) % 60),
    total,
  }
}

export default function Countdown({
  targetDate,
  onExpire,
  showLabels = true,
  className = '',
  ...props
}: CountdownProps) {
  const target = typeof targetDate === 'string' ? new Date(targetDate) : targetDate
  const [timeLeft, setTimeLeft] = useState(() => calculateTimeLeft(target))

  useEffect(() => {
    const timer = setInterval(() => {
      const newTimeLeft = calculateTimeLeft(target)
      setTimeLeft(newTimeLeft)

      if (newTimeLeft.total <= 0) {
        clearInterval(timer)
        onExpire?.()
      }
    }, 1000)

    return () => clearInterval(timer)
  }, [target, onExpire])

  if (timeLeft.total <= 0) {
    return (
      <span className={`text-sm font-medium text-red-600 ${className}`} {...props}>
        Expired
      </span>
    )
  }

  const segments = [
    { value: timeLeft.days, label: 'd' },
    { value: timeLeft.hours, label: 'h' },
    { value: timeLeft.minutes, label: 'm' },
    { value: timeLeft.seconds, label: 's' },
  ]

  return (
    <div className={`flex items-center gap-1 ${className}`} {...props}>
      {segments.map(({ value, label }, i) => (
        <React.Fragment key={label}>
          {i > 0 && <span className="text-slate-400">:</span>}
          <span className="font-mono text-sm font-medium text-slate-700">
            {String(value).padStart(2, '0')}
          </span>
          {showLabels && <span className="text-xs text-slate-500">{label}</span>}
        </React.Fragment>
      ))}
    </div>
  )
}
