'use client';

import { CheckCircle, Circle, Loader2, Wrench } from 'lucide-react';

export type FSMStage =
  | 'INTRODUCTION'
  | 'JOB_DESCRIPTION'
  | 'SALARY_NEGOTIATION'
  | 'INTERVIEWING'
  | 'OFFER'
  | 'COMPLETED';

interface ToolCall {
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  cached?: boolean;
}

interface FlowPanelProps {
  currentStage: FSMStage;
  toolCalls?: ToolCall[];
  className?: string;
}

const STAGES: { key: FSMStage; label: string }[] = [
  { key: 'INTRODUCTION', label: 'Introduction' },
  { key: 'JOB_DESCRIPTION', label: 'Job Description' },
  { key: 'SALARY_NEGOTIATION', label: 'Salary Negotiation' },
  { key: 'INTERVIEWING', label: 'Interview' },
  { key: 'OFFER', label: 'Offer' },
  { key: 'COMPLETED', label: 'Completed' },
];

const STAGE_ORDER: FSMStage[] = STAGES.map(s => s.key);

function StageProgress({ currentStage }: { currentStage: FSMStage }) {
  const currentIdx = STAGE_ORDER.indexOf(currentStage);

  return (
    <div className="space-y-2">
      {STAGES.map((stage, idx) => {
        const isCompleted = idx < currentIdx;
        const isCurrent = idx === currentIdx;
        const isFuture = idx > currentIdx;

        return (
          <div key={stage.key} className="flex items-center gap-3">
            {isCompleted && <CheckCircle className="w-5 h-5 text-green-500 shrink-0" />}
            {isCurrent && <Loader2 className="w-5 h-5 text-blue-500 animate-spin shrink-0" />}
            {isFuture && <Circle className="w-5 h-5 text-gray-300 shrink-0" />}
            <span className={`text-sm ${
              isCompleted ? 'text-green-700 line-through' :
              isCurrent ? 'text-blue-700 font-medium' :
              'text-gray-400'
            }`}>
              {stage.label}
            </span>
          </div>
        );
      })}
    </div>
  );
}

function ToolCallList({ toolCalls }: { toolCalls: ToolCall[] }) {
  if (toolCalls.length === 0) return null;

  return (
    <div className="space-y-2">
      {toolCalls.map((tc, idx) => (
        <div key={idx} className="flex items-center gap-2 text-sm">
          <Wrench className="w-4 h-4 text-gray-400 shrink-0" />
          <span className="flex-1 text-gray-700 truncate">{tc.name}</span>
          {tc.status === 'completed' && <CheckCircle className="w-4 h-4 text-green-500" />}
          {tc.status === 'running' && <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />}
          {tc.status === 'failed' && <Circle className="w-4 h-4 text-red-500" />}
          {tc.cached && (
            <span className="px-1.5 py-0.5 text-xs bg-amber-100 text-amber-700 rounded">cached</span>
          )}
        </div>
      ))}
    </div>
  );
}

export function FlowPanel({ currentStage, toolCalls = [], className = '' }: FlowPanelProps) {
  return (
    <div className={`space-y-6 ${className}`}>
      {/* FSM Stage Badge */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-2">Stage</h3>
        <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800">
          {currentStage.replace(/_/g, ' ')}
        </span>
      </div>

      {/* Progress */}
      <div>
        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">Progress</h3>
        <StageProgress currentStage={currentStage} />
      </div>

      {/* Tool Calls */}
      {toolCalls.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">Tool Calls</h3>
          <ToolCallList toolCalls={toolCalls} />
        </div>
      )}
    </div>
  );
}
