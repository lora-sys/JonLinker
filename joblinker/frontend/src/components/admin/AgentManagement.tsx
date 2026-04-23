'use client';

import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';

interface Agent {
  id: string;
  type: string;
  status: string;
  created_at: string;
  user?: {
    email?: string;
  };
}

interface AgentManagementProps {
  agents: Agent[];
  onPauseAgent?: (agentId: string) => void;
  onResumeAgent?: (agentId: string) => void;
  onViewDetails?: (agentId: string) => void;
}

export function AgentManagement({
  agents,
  onPauseAgent,
  onResumeAgent,
  onViewDetails,
}: AgentManagementProps) {
  const statusColors: Record<string, string> = {
    active: 'bg-green-100 text-green-800',
    paused: 'bg-yellow-100 text-yellow-800',
    inactive: 'bg-gray-100 text-gray-800',
  };

  const typeLabels: Record<string, string> = {
    seeker: 'Job Seeker',
    recruiter: 'Recruiter',
  };

  return (
    <Card variant="outlined" padding="none">
      <div className="p-4 border-b border-gray-200 flex items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-900">Agent Management</h3>
        {onPauseAgent && (
          <Button variant="secondary" size="sm">
            Pause All
          </Button>
        )}
      </div>
      <div className="divide-y divide-gray-100">
        {agents.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            No agents found
          </div>
        ) : (
          agents.map((agent) => (
            <div key={agent.id} className="p-4 hover:bg-gray-50 transition-colors">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-gray-900">
                      {typeLabels[agent.type] || agent.type}
                    </span>
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColors[agent.status] || 'bg-gray-100 text-gray-800'}`}>
                      {agent.status}
                    </span>
                  </div>
                  <p className="text-sm text-gray-500 mt-1">
                    {agent.user?.email || 'No email'}
                  </p>
                  <p className="text-xs text-gray-400 mt-1">
                    Created {new Date(agent.created_at).toLocaleDateString()}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  {agent.status === 'active' && onPauseAgent && (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => onPauseAgent(agent.id)}
                    >
                      Pause
                    </Button>
                  )}
                  {agent.status === 'paused' && onResumeAgent && (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => onResumeAgent(agent.id)}
                    >
                      Resume
                    </Button>
                  )}
                  {onViewDetails && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => onViewDetails(agent.id)}
                    >
                      Details
                    </Button>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </Card>
  );
}
