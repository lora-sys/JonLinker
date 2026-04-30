'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth';
import { useRole } from '@/hooks/useRole';
import JobForm from '@/components/ui/job-form';
import { BackgroundBeams } from '@/components/ui/background-beams';

export default function NewJobPage() {
  const router = useRouter();
  const role = useRole();
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    // Auth is now handled by middleware via cookie
    // Only need role check client-side
    setIsReady(true);

    if (role !== 'recruiter') {
      router.replace('/dashboard');
    }
  }, [role, router]);

  if (!isReady) {
    return (
      <div className="min-h-screen bg-zinc-950 flex items-center justify-center">
        <div className="animate-pulse text-cyan-400">Loading...</div>
      </div>
    );
  }

  if (role !== 'recruiter') {
    return null;
  }

  return (
    <div className="min-h-screen bg-zinc-950 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-cyan-950/20 via-zinc-950 to-zinc-950 pointer-events-none" />
      <BackgroundBeams />
      <div className="relative z-10 flex items-center justify-center min-h-screen px-4 py-12">
        <JobForm />
      </div>
    </div>
  );
}