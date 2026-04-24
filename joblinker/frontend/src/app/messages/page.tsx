'use client';

import Card from '@/components/ui/Card';
import ConnectionStatus from '@/components/ui/ConnectionStatus';
import TypingIndicator from '@/components/ui/TypingIndicator';
import UnreadBadge from '@/components/ui/UnreadBadge';

export default function MessagesPage() {
  const connectionState = 'connected' as const;
  const isTyping = false;
  const unreadCount = 3;

  return (
    <div className="max-w-4xl mx-auto px-4 py-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-slate-900" style={{ fontFamily: 'Plus Jakarta Sans, sans-serif' }}>Messages</h1>
        <div className="flex items-center gap-3">
          <ConnectionStatus state={connectionState} />
          <UnreadBadge count={unreadCount} />
        </div>
      </div>

      {isTyping && (
        <Card className="p-4 mb-4">
          <TypingIndicator name="Agent" />
        </Card>
      )}

      <Card className="p-10 text-center">
        <div className="w-14 h-14 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg className="w-7 h-7 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        </div>
        <h2 className="text-lg font-semibold text-slate-900 mb-2">No messages yet</h2>
        <p className="text-slate-500">Start a conversation with a match to see messages here</p>
      </Card>
    </div>
  );
}
