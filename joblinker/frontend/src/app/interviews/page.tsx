'use client';

import Card from '@/components/ui/Card';
import Badge from '@/components/ui/Badge';
import Countdown from '@/components/ui/Countdown';

type InterviewType = 'video' | 'phone' | 'onsite';
type InterviewStatus = 'scheduled' | 'completed' | 'cancelled';

interface Interview {
  id: string;
  type: InterviewType;
  status: InterviewStatus;
  scheduledAt: Date;
  company: string;
  role: string;
}

const MOCK_INTERVIEWS: Interview[] = [
  {
    id: '1',
    type: 'video',
    status: 'scheduled',
    scheduledAt: new Date(Date.now() + 2 * 24 * 60 * 60 * 1000),
    company: 'TechCorp',
    role: 'Senior React Developer',
  },
  {
    id: '2',
    type: 'onsite',
    status: 'scheduled',
    scheduledAt: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000),
    company: 'StartupXYZ',
    role: 'Backend Engineer',
  },
];

const typeIcons: Record<InterviewType, React.ReactNode> = {
  video: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
    </svg>
  ),
  phone: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
    </svg>
  ),
  onsite: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
    </svg>
  ),
};

const typeColors: Record<InterviewType, string> = {
  video: 'bg-blue-100 text-blue-600',
  phone: 'bg-green-100 text-green-600',
  onsite: 'bg-purple-100 text-purple-600',
};

export default function InterviewsPage() {
  const scheduledInterviews = MOCK_INTERVIEWS.filter(i => i.status === 'scheduled');

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold text-slate-900 mb-6">Interviews</h1>

        {scheduledInterviews.length === 0 ? (
          <Card className="p-10 text-center bg-white/80 backdrop-blur-xl border border-white/20">
            <div className="w-14 h-14 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-7 h-7 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <h2 className="text-lg font-semibold text-slate-900 mb-2">No interviews scheduled</h2>
            <p className="text-slate-500">Interviews will appear here once matches progress to that stage</p>
          </Card>
        ) : (
          <div className="space-y-6">
            {/* Timeline View */}
            <div className="relative">
              <div className="absolute left-6 top-0 bottom-0 w-0.5 bg-slate-200" />
              <div className="space-y-6">
                {scheduledInterviews.map((interview, index) => (
                  <div key={interview.id} className="relative flex gap-4">
                    {/* Timeline Dot */}
                    <div className={`relative z-10 w-12 h-12 rounded-full flex items-center justify-center ${typeColors[interview.type]}`}>
                      {typeIcons[interview.type]}
                    </div>

                    {/* Content Card */}
                    <Card hover className="flex-1 p-5 bg-white/80 backdrop-blur-xl border border-white/20">
                      <div className="flex items-start justify-between mb-3">
                        <div>
                          <h3 className="font-semibold text-slate-900">{interview.role}</h3>
                          <p className="text-sm text-slate-600">{interview.company}</p>
                        </div>
                        <Badge status="pending">{interview.status}</Badge>
                      </div>

                      {/* Countdown Timer */}
                      <div className="flex items-center gap-2 text-sm text-slate-500 mb-3">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <Countdown targetDate={interview.scheduledAt} showLabels />
                      </div>

                      {/* Interview Type Label */}
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-full bg-slate-100 text-slate-600 capitalize">
                        {typeIcons[interview.type]}
                        {interview.type} Interview
                      </span>
                    </Card>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}