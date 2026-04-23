'use client';

interface MessageBubbleProps {
  content: string;
  intent: string;
  timestamp: string;
  sender?: 'self' | 'remote';
}

export function MessageBubble({ content, intent, timestamp, sender = 'remote' }: MessageBubbleProps) {
  const formatTime = (ts: string) => {
    try {
      return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  const intentColors: Record<string, string> = {
    INTRODUCTION: 'bg-blue-50 border-blue-200',
    INTEREST: 'bg-green-50 border-green-200',
    NEGOTIATION: 'bg-yellow-50 border-yellow-200',
    OFFER: 'bg-purple-50 border-purple-200',
    ACCEPT: 'bg-green-100 border-green-300',
    DECLINE: 'bg-red-50 border-red-200',
    SCHEDULE: 'bg-blue-50 border-blue-200',
    CONFIRM: 'bg-green-100 border-green-300',
    INQUIRY: 'bg-gray-50 border-gray-200',
  };

  const alignment = sender === 'self' ? 'justify-end' : 'justify-start';

  return (
    <div className={`flex ${alignment}`}>
      <div className={`max-w-[80%] p-3 rounded-lg border ${
        intentColors[intent] || 'bg-gray-50 border-gray-200'
      } ${sender === 'self' ? 'bg-blue-100 border-blue-300' : ''}`}>
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1">
            <p className="text-gray-900">{content}</p>
            {intent && (
              <span className="inline-block mt-1 px-2 py-0.5 text-xs font-medium bg-white/50 rounded">
                {intent}
              </span>
            )}
          </div>
          <span className="text-xs text-gray-400 whitespace-nowrap">{formatTime(timestamp)}</span>
        </div>
      </div>
    </div>
  );
}