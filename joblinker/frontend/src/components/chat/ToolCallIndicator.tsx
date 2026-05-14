'use client';

import { useState } from 'react';
import { Loader2, CheckCircle2, AlertCircle, ChevronDown, ChevronUp } from 'lucide-react';
import type { ToolCall } from '@/types/ai';

interface ToolCallIndicatorProps {
  tool: ToolCall;
}

export function ToolCallIndicator({ tool }: ToolCallIndicatorProps) {
  const [expanded, setExpanded] = useState(false);

  const statusConfig = {
    pending: { icon: Loader2, color: 'text-amber-600', bg: 'bg-amber-50', border: 'border-amber-200', spin: true },
    in_progress: { icon: Loader2, color: 'text-blue-600', bg: 'bg-blue-50', border: 'border-blue-200', spin: true },
    done: { icon: CheckCircle2, color: 'text-green-600', bg: 'bg-green-50', border: 'border-green-200', spin: false },
    error: { icon: AlertCircle, color: 'text-red-600', bg: 'bg-red-50', border: 'border-red-200', spin: false },
  };

  const config = statusConfig[tool.status];
  const Icon = config.icon;

  const toolNameLabels: Record<string, string> = {
    job_query: 'Searching for jobs',
    offer_create: 'Creating offer',
    interview_schedule: 'Scheduling interview',
    candidate_search: 'Searching candidates',
    salary_analysis: 'Analyzing salary',
  };

  const label = toolNameLabels[tool.toolName] || `Calling tool: ${tool.toolName}`;

  return (
    <div className={`rounded-lg border ${config.border} ${config.bg} overflow-hidden`}>
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-2 px-3 py-2 text-sm"
        aria-expanded={expanded}
      >
        <Icon className={`w-4 h-4 ${config.color} ${config.spin ? 'animate-spin' : ''}`} aria-hidden="true" />
        <span className={`${config.color} font-medium flex-1 text-left animate-pulse`}>
          {tool.status === 'done' ? `${label} - Complete` : `${label}...`}
        </span>
        {tool.status === 'in_progress' && (
          <div className="w-16 h-1.5 bg-blue-100 rounded-full overflow-hidden">
            <div className="h-full bg-blue-500 animate-pulse origin-left" style={{ width: '60%' }} />
          </div>
        )}
        {expanded ? (
          <ChevronUp className="w-4 h-4 text-gray-400" />
        ) : (
          <ChevronDown className="w-4 h-4 text-gray-400" />
        )}
      </button>

      {expanded && (
        <div className="px-3 pb-3 border-t border-gray-200/50">
          <div className="mt-2 text-xs text-gray-600">
            <p className="font-medium mb-1">Arguments:</p>
            <pre className="bg-white/60 rounded p-2 overflow-x-auto">
              {JSON.stringify(tool.args, null, 2)}
            </pre>
          </div>

          {tool.status === 'done' && tool.result !== undefined && (
            <div className="mt-2 text-xs text-gray-600">
              <p className="font-medium mb-1">Result:</p>
              <pre className="bg-white/60 rounded p-2 overflow-x-auto">
                {JSON.stringify(tool.result, null, 2)}
              </pre>
            </div>
          )}

          {tool.status === 'error' && tool.error && (
            <div className="mt-2 text-xs text-red-600">
              <p className="font-medium mb-1">Error:</p>
              <p className="bg-white/60 rounded p-2">{tool.error}</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
