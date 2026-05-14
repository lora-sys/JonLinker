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

  const intentGradients: Record<string, string> = {
    INTRODUCTION: 'bg-gradient-to-br from-blue-50 to-blue-100 border-blue-200',
    INTEREST: 'bg-gradient-to-br from-green-50 to-green-100 border-green-200',
    NEGOTIATION: 'bg-gradient-to-br from-yellow-50 to-amber-100 border-yellow-200',
    OFFER: 'bg-gradient-to-br from-purple-50 to-violet-100 border-purple-200',
    ACCEPT: 'bg-gradient-to-br from-green-100 to-emerald-100 border-green-300',
    DECLINE: 'bg-gradient-to-br from-red-50 to-rose-100 border-red-200',
    SCHEDULE: 'bg-gradient-to-br from-blue-50 to-cyan-100 border-blue-200',
    CONFIRM: 'bg-gradient-to-br from-green-100 to-emerald-100 border-green-300',
    INQUIRY: 'bg-gradient-to-br from-gray-50 to-slate-100 border-gray-200',
  };

  const alignment = sender === 'self' ? 'justify-end' : 'justify-start';
  const senderGradient = sender === 'self' ? 'bg-gradient-to-br from-blue-500 to-blue-600 border-blue-400' : '';

  return (
    <div className={`flex ${alignment} animate-fadeIn`}>
      <div className={`max-w-[80%] p-3 rounded-xl border shadow-sm transition-all duration-200 hover:shadow-md ${
        senderGradient || intentGradients[intent] || 'bg-gradient-to-br from-gray-50 to-gray-100 border-gray-200'
      }`}>
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1">
            <p className="text-gray-900">{content}</p>
            {intent && (
              <span className="inline-block mt-1 px-2 py-0.5 text-xs font-medium bg-white/60 rounded">
                {intent}
              </span>
            )}
          </div>
          <span className="text-xs text-gray-500 whitespace-nowrap">{formatTime(timestamp)}</span>
        </div>
      </div>
    </div>
  );
}