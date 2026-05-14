'use client';

import { useCallback, useRef, useEffect, useState } from 'react';
import { Send, Square, RotateCcw } from 'lucide-react';

interface TextareaInputProps {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
  onStop?: () => void;
  onRetry?: () => void;
  disabled?: boolean;
  isStreaming?: boolean;
  error?: string | null;
}

export function TextareaInput({
  value,
  onChange,
  onSubmit,
  onStop,
  onRetry,
  disabled = false,
  isStreaming = false,
  error = null,
}: TextareaInputProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [isSending, setIsSending] = useState(false);

  const autoResize = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = 'auto';
    const lineHeight = 24;
    const maxLines = 5;
    const maxHeight = lineHeight * maxLines;
    el.style.height = `${Math.min(el.scrollHeight, maxHeight)}px`;
  }, []);

  useEffect(() => {
    autoResize();
  }, [value, autoResize]);

  const handleSubmit = useCallback(() => {
    if (!value.trim() || disabled) return;
    setIsSending(true);
    onSubmit();
    setTimeout(() => setIsSending(false), 300);
  }, [value, disabled, onSubmit]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        if (!disabled && value.trim()) {
          handleSubmit();
        }
      }
      if (e.key === 'Escape' && isStreaming) {
        e.preventDefault();
        onStop?.();
      }
    },
    [disabled, value, handleSubmit, isStreaming, onStop]
  );

  return (
    <div className="border-t border-gray-200 bg-white p-3">
      {error && (
        <div className="mb-2 p-2 bg-red-50 text-red-700 text-sm rounded-lg flex items-center justify-between">
          <span>{error}</span>
          {onRetry && (
            <button
              onClick={onRetry}
              className="flex items-center gap-1 px-2 py-1 text-xs font-medium text-red-700 bg-red-100 rounded-md hover:bg-red-200"
            >
              <RotateCcw className="w-3 h-3" />
              Retry
            </button>
          )}
        </div>
      )}

      <div className="flex items-end gap-2">
        <textarea
          ref={textareaRef}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Type a message... (Shift+Enter for new line)"
          rows={1}
          className={`flex-1 px-3 py-2 border border-gray-300 rounded-lg resize-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50 text-sm leading-6 transition-all duration-200 ${isSending ? 'scale-98 opacity-80' : ''}`}
          disabled={disabled}
          aria-label="Message input"
        />

        {isStreaming ? (
          <button
            onClick={onStop}
            className="p-2.5 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-all duration-200 shrink-0 animate-pulse"
            aria-label="Stop generating"
          >
            <Square className="w-4 h-4" />
          </button>
        ) : (
          <button
            onClick={handleSubmit}
            disabled={!value.trim() || disabled}
            className={`p-2.5 text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shrink-0 ${isSending ? 'bg-blue-700 scale-95' : 'bg-blue-600 hover:bg-blue-700'}`}
            aria-label="Send message"
          >
            <Send className={`w-4 h-4 ${isSending ? 'animate-bounce' : ''}`} />
          </button>
        )}
      </div>
    </div>
  );
}
