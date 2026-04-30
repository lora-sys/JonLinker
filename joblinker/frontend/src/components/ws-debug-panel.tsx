'use client';

import React, { useState, useCallback } from 'react';
import { useWebSocket, WSConnectionStatus, WSMessage } from '@/hooks/useWebSocket';
import { buildWSUrl } from '@/lib/websocket';
import { getAuthToken } from '@/lib/api-utils';
import { cn } from '@/lib/utils';
import { RefreshCw, Send, Wifi, WifiOff, Loader2 } from 'lucide-react';

interface WSDebugPanelProps {
  token?: string;
  className?: string;
}

interface LogEntry {
  id: number;
  timestamp: Date;
  direction: '→' | '←';
  message: WSMessage;
}

const statusConfig: Record<WSConnectionStatus, { label: string; color: string; bg: string }> = {
  Connecting: { label: 'Connecting', color: 'text-yellow-400', bg: 'bg-yellow-500/10 border-yellow-500/30' },
  Connected: { label: 'Connected', color: 'text-green-400', bg: 'bg-green-500/10 border-green-500/30' },
  Disconnected: { label: 'Disconnected', color: 'text-red-400', bg: 'bg-red-500/10 border-red-500/30' },
  Reconnecting: { label: 'Reconnecting', color: 'text-orange-400', bg: 'bg-orange-500/10 border-orange-500/30' },
};

export default function WSDebugPanel({ token, className }: WSDebugPanelProps) {
  const authToken = token || getAuthToken();
  const [wsUrl] = useState(() => authToken ? buildWSUrl(authToken) : '');
  const [inputMessage, setInputMessage] = useState('');
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [logIdCounter, setLogIdCounter] = useState(0);

  const handleMessage = useCallback((msg: WSMessage) => {
    setLogs(prev => [...prev, {
      id: logIdCounter,
      timestamp: new Date(),
      direction: '←' as const,
      message: msg,
    }]);
    setLogIdCounter(c => c + 1);
  }, [logIdCounter]);

  const handleStatusChange = useCallback((status: WSConnectionStatus) => {
    console.log('[WS] Status:', status);
  }, []);

  const { status, send, reconnect } = useWebSocket({
    url: wsUrl,
    onMessage: handleMessage,
    onStatusChange: handleStatusChange,
    autoConnect: !!authToken,
  });

  const handleSend = () => {
    if (!inputMessage.trim() || status !== 'Connected') return;

    const msg: WSMessage = {
      type: 'message',
      content: inputMessage.trim(),
      sender: 'debug-panel',
      timestamp: new Date().toISOString(),
    };

    send(msg);
    setLogs(prev => [...prev, {
      id: logIdCounter,
      timestamp: new Date(),
      direction: '→' as const,
      message: msg,
    }]);
    setLogIdCounter(c => c + 1);
    setInputMessage('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  const statusInfo = statusConfig[status];

  return (
    <div className={cn('flex flex-col h-full', className)}>
      {/* Status Header */}
      <div className="flex items-center justify-between p-4 border-b border-zinc-800">
        <div className="flex items-center gap-3">
          {status === 'Connected' ? (
            <Wifi className="h-5 w-5 text-green-400" />
          ) : status === 'Connecting' || status === 'Reconnecting' ? (
            <Loader2 className="h-5 w-5 text-yellow-400 animate-spin" />
          ) : (
            <WifiOff className="h-5 w-5 text-red-400" />
          )}
          <span className={cn('text-sm font-medium', statusInfo.color)}>
            {statusInfo.label}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={reconnect}
            disabled={status === 'Connecting'}
            className="p-2 rounded-lg hover:bg-zinc-800 transition-colors disabled:opacity-50"
            title="Reconnect"
          >
            <RefreshCw className={cn('h-4 w-4 text-zinc-400', status === 'Reconnecting' && 'animate-spin')} />
          </button>
        </div>
      </div>

      {/* Connection URL (for debugging) */}
      <div className="px-4 py-2 bg-zinc-900/50 border-b border-zinc-800">
        <p className="text-xs text-zinc-500 truncate font-mono">
          {wsUrl.split('?')[0]}?token=***
        </p>
      </div>

      {/* Log Area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-2 min-h-[200px]">
        {logs.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <p className="text-zinc-500 text-sm">
              {status === 'Connected' ? 'Send a message to see it here' : 'Not connected'}
            </p>
          </div>
        ) : (
          logs.map((entry) => (
            <div
              key={entry.id}
              className={cn(
                'flex items-start gap-3 p-3 rounded-lg',
                entry.direction === '→' ? 'bg-cyan-500/5 border border-cyan-500/10' : 'bg-zinc-800/50 border border-zinc-700/50'
              )}
            >
              <span className={cn(
                'text-xs font-mono font-bold w-6',
                entry.direction === '→' ? 'text-cyan-400' : 'text-zinc-400'
              )}>
                {entry.direction}
              </span>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs text-zinc-500 font-mono">
                    {formatTime(entry.timestamp)}
                  </span>
                  <span className={cn(
                    'px-1.5 py-0.5 rounded text-xs font-medium',
                    entry.message.type === 'ping' ? 'bg-green-500/10 text-green-400' :
                    entry.message.type === 'pong' ? 'bg-blue-500/10 text-blue-400' :
                    entry.message.type === 'message' ? 'bg-purple-500/10 text-purple-400' :
                    'bg-zinc-700/50 text-zinc-400'
                  )}>
                    {entry.message.type}
                  </span>
                </div>
                <pre className="text-xs text-zinc-300 font-mono whitespace-pre-wrap break-all">
                  {JSON.stringify(entry.message, null, 0)}
                </pre>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Input Area */}
      <div className="p-4 border-t border-zinc-800 space-y-3">
        <div className="flex items-center gap-2">
          <input
            type="text"
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={status === 'Connected' ? 'Type a message...' : 'Not connected'}
            disabled={status !== 'Connected'}
            className="flex-1 h-10 px-4 rounded-lg bg-zinc-900 border border-zinc-700 text-white text-sm placeholder:text-zinc-500 focus:border-cyan-500/50 focus:outline-none focus:ring-2 focus:ring-cyan-500/10 disabled:opacity-50"
          />
          <button
            onClick={handleSend}
            disabled={status !== 'Connected' || !inputMessage.trim()}
            className="h-10 px-4 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-black text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <Send className="h-4 w-4" />
            Send
          </button>
        </div>
        {status !== 'Connected' && (
          <button
            onClick={reconnect}
            className="w-full h-10 rounded-lg border border-zinc-700 hover:border-zinc-600 text-zinc-400 text-sm font-medium transition-colors flex items-center justify-center gap-2"
          >
            <RefreshCw className="h-4 w-4" />
            Try to Reconnect
          </button>
        )}
      </div>
    </div>
  );
}