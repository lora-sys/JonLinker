'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/ui/Card';

interface Job {
  id: string;
  title: string;
  company: string;
  location: string;
  skills: string[];
  salary: string;
  type: string;
}

const MOCK_JOBS: Job[] = [
  { id: '1', title: 'Senior React Developer', company: 'TechCorp', location: 'Remote', skills: ['react', 'typescript', 'node'], salary: '$120k-$150k', type: 'Full-time' },
  { id: '2', title: 'Backend Engineer', company: 'StartupXYZ', location: 'San Francisco', skills: ['go', 'python', 'sql'], salary: '$130k-$160k', type: 'Full-time' },
  { id: '3', title: 'DevOps Engineer', company: 'CloudBase', location: 'Remote', skills: ['kubernetes', 'docker', 'aws'], salary: '$110k-$140k', type: 'Contract' },
];

export default function JobsPage() {
  const [search, setSearch] = useState('');
  const [skillFilter, setSkillFilter] = useState('');
  const [sortBy, setSortBy] = useState<'title' | 'company'>('title');

  const allSkills = useMemo(() => {
    const skills = new Set<string>();
    MOCK_JOBS.forEach(job => job.skills.forEach(s => skills.add(s)));
    return Array.from(skills).sort();
  }, []);

  const filteredJobs = useMemo(() => {
    return MOCK_JOBS
      .filter(job => {
        const matchesSearch = job.title.toLowerCase().includes(search.toLowerCase()) ||
                            job.company.toLowerCase().includes(search.toLowerCase());
        const matchesSkill = !skillFilter || job.skills.includes(skillFilter);
        return matchesSearch && matchesSkill;
      })
      .sort((a, b) => a[sortBy].localeCompare(b[sortBy]));
  }, [search, skillFilter, sortBy]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-900">Job Board</h1>
          <p className="text-slate-600 mt-1">Discover opportunities that match your skills</p>
        </div>

        {/* Sticky Search & Filter Bar */}
        <Card className="p-4 mb-6 sticky top-4 z-10 bg-white/80 backdrop-blur-xl border border-white/20">
          <div className="flex flex-col md:flex-row gap-3">
            {/* Search Input */}
            <div className="flex-1 relative">
              <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <input
                type="text"
                placeholder="Search jobs..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full pl-10 pr-4 py-2.5 bg-slate-50/50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent cursor-text transition-all duration-200"
              />
            </div>

            {/* Skill Filter Dropdown */}
            <select
              value={skillFilter}
              onChange={(e) => setSkillFilter(e.target.value)}
              className="px-4 py-2.5 bg-slate-50/50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer transition-all duration-200"
            >
              <option value="">All Skills</option>
              {allSkills.map(skill => (
                <option key={skill} value={skill}>{skill}</option>
              ))}
            </select>

            {/* Sort Dropdown */}
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as 'title' | 'company')}
              className="px-4 py-2.5 bg-slate-50/50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer transition-all duration-200"
            >
              <option value="title">Sort by Title</option>
              <option value="company">Sort by Company</option>
            </select>
          </div>
        </Card>

        {/* Job Listings */}
        {filteredJobs.length === 0 ? (
          <Card className="p-10 text-center bg-white/80 backdrop-blur-xl border border-white/20">
            <div className="w-16 h-16 bg-gradient-to-br from-blue-100 to-sky-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
            </div>
            <h2 className="text-lg font-semibold text-slate-900 mb-2">No jobs found</h2>
            <p className="text-slate-500">Try adjusting your search or filter criteria</p>
          </Card>
        ) : (
          <div className="space-y-4">
            {filteredJobs.map((job) => (
              <Card key={job.id} hover className="p-5 bg-white/80 backdrop-blur-xl border border-white/20 cursor-pointer">
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-semibold text-slate-900">{job.title}</h3>
                    <p className="text-sm text-slate-600 mt-1">{job.company} • {job.location}</p>
                    <div className="flex gap-2 mt-2">
                      {job.skills.map(skill => (
                        <span key={skill} className="px-2.5 py-1 text-xs bg-blue-50 text-blue-600 rounded-full font-medium">
                          {skill}
                        </span>
                      ))}
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-semibold text-slate-900">{job.salary}</p>
                    <p className="text-xs text-slate-500 mt-1">{job.type}</p>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
