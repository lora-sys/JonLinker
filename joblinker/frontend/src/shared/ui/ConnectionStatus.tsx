'use client'

import React from 'react'

type ConnectionState = 'connected' | 'connecting' | 'disconnected' | 'error' | 'reconnecting' | 'failed'

interface ConnectionStatusProps extends React.HTMLAttributes<HTMLDivElement> {
  state?: ConnectionState
  showLabel?: boolean
}

const stateConfig: Record<ConnectionState, { color: string, label: string, animate?: boolean }> = {
  connected: { color: 'bg-green-500', label: 'Connected' },
  connecting: { color: 'bg-yellow-500', label: 'Connecting...', animate: true },
  disconnected: { color: 'bg-slate-400', label: 'Disconnected' },
  error: { color: 'bg-red-500', label: 'Connection error' },
  reconnecting: { color: 'bg-amber-500', label: 'Reconnecting...', animate: true },
  failed: { color: 'bg-red-500', label: 'Connection failed' },
}

export default function ConnectionStatus({
  state = 'disconnected',
  showLabel = true,
  className = '',
  ...props
}: ConnectionStatusProps) {
  const config = stateConfig[state]

  return (
    <div className={`flex items-center gap-2 ${className}`} {...props}>
      <span className="relative flex h-3 w-3">
        {config.animate && (
          <span
            className={`absolute inline-flex h-full w-full rounded-full opacity-75 ${config.color} animate-ping`}
          />
        )}
        <span className={`relative inline-flex rounded-full h-3 w-3 ${config.color}`} />
      </span>
      {showLabel && (
        <span className="text-xs font-medium text-slate-600">{config.label}</span>
      )}
    </div>
  )
}
