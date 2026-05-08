'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Sparkles, ArrowLeft, Loader2, FileText, Check, X } from 'lucide-react';
import { apiClient } from '@/lib/api_client';

export default function ResumeGeneratePage() {
  const router = useRouter();
  const [step, setStep] = useState<'input' | 'generating' | 'result'>('input');
  const [userInfo, setUserInfo] = useState('');
  const [resume, setResume] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  const [notification, setNotification] = useState<string | null>(null);

  const handleGenerate = async () => {
    if (!userInfo.trim()) {
      setError('Please provide some information about yourself');
      return;
    }

    setStep('generating');
    setError(null);

    try {
      const data = await apiClient.post<{ resume: any; raw: string }>('/api/resumes/generate', {
        user_info: userInfo,
      });
      setResume(data.resume || JSON.parse(data.raw || '{}'));
      setStep('result');
    } catch (err) {
      setError('Failed to generate resume. Please try again.');
      setStep('input');
    }
  };

  const handleSaveToStorage = async () => {
    if (!resume) return;

    try {
      // Import privacy storage
      const { privacyStorage, initPrivacyStore } = await import('@/lib/privacy');
      await initPrivacyStore();

      const password = prompt('Enter a password to encrypt your resume (min 8 characters):');
      if (!password || password.length < 8) {
        setNotification('Password must be at least 8 characters');
        setTimeout(() => setNotification(null), 3000);
        return;
      }

      const id = `ai-resume-${Date.now()}`;
      const resumeText = JSON.stringify(resume, null, 2);
      await privacyStorage.storeResume(id, resumeText, password);

      setNotification('Resume saved to your encrypted local storage!');
      setTimeout(() => {
        setNotification(null);
        router.push('/resume');
      }, 1500);
    } catch (err) {
      setNotification('Failed to save resume');
      setTimeout(() => setNotification(null), 3000);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Toast notification */}
        {notification && (
          <div className="fixed top-4 right-4 z-50 flex items-center gap-2 px-4 py-3 bg-blue-600 text-white rounded-lg shadow-lg">
            <span className="text-sm font-medium">{notification}</span>
            <button onClick={() => setNotification(null)} className="p-0.5 hover:bg-blue-700 rounded">
              <X className="w-4 h-4" />
            </button>
          </div>
        )}

        <button
          onClick={() => router.back()}
          className="flex items-center gap-2 text-slate-600 hover:text-slate-900 mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>

        <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3 mb-6">
          <Sparkles className="w-8 h-8 text-amber-500" />
          AI Resume Generator
        </h1>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
            {error}
          </div>
        )}

        {step === 'input' && (
          <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-6">
            <p className="text-slate-600 mb-4">
              Enter some information about yourself and our AI will generate a professional resume for you.
            </p>

            <textarea
              value={userInfo}
              onChange={(e) => setUserInfo(e.target.value)}
              placeholder="Example:
Name: John Doe
Email: john@example.com
Phone: (555) 123-4567

Experience:
- Software Engineer at Tech Corp (2020-2024)
  - Built REST APIs using Go and Python
  - Led team of 5 engineers

Education:
- BS Computer Science, MIT (2016-2020)

Skills: Go, Python, JavaScript, PostgreSQL, Docker"
              className="w-full h-64 px-4 py-3 border border-slate-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
            />

            <button
              onClick={handleGenerate}
              className="mt-4 w-full py-3 bg-gradient-to-r from-amber-500 to-orange-500 text-white rounded-xl font-semibold hover:from-amber-600 hover:to-orange-600 transition-all duration-200 shadow-lg shadow-amber-500/25 flex items-center justify-center gap-2"
            >
              <Sparkles className="w-5 h-5" />
              Generate Resume with AI
            </button>
          </div>
        )}

        {step === 'generating' && (
          <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-12 text-center">
            <Loader2 className="w-12 h-12 text-blue-600 animate-spin mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-slate-900 mb-2">Generating your resume...</h2>
            <p className="text-slate-500">Our AI is crafting a professional resume based on your information</p>
          </div>
        )}

        {step === 'result' && resume && (
          <div className="space-y-6">
            <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-semibold text-slate-900 flex items-center gap-2">
                  <FileText className="w-5 h-5 text-green-600" />
                  Your Generated Resume
                </h2>
                <span className="flex items-center gap-1 text-green-600 text-sm">
                  <Check className="w-4 h-4" />
                  AI Generated
                </span>
              </div>

              <div className="bg-slate-50 rounded-xl p-6 space-y-6">
                {resume.summary && (
                  <div>
                    <h3 className="font-semibold text-slate-900 mb-2">Summary</h3>
                    <p className="text-slate-700">{resume.summary}</p>
                  </div>
                )}

                {resume.skills && (
                  <div>
                    <h3 className="font-semibold text-slate-900 mb-2">Skills</h3>
                    <div className="flex flex-wrap gap-2">
                      {resume.skills.map((skill: string, i: number) => (
                        <span key={i} className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm">
                          {skill}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {resume.experience && (
                  <div>
                    <h3 className="font-semibold text-slate-900 mb-2">Experience</h3>
                    <div className="space-y-4">
                      {resume.experience.map((exp: any, i: number) => (
                        <div key={i} className="border-l-2 border-blue-200 pl-4">
                          <p className="font-medium text-slate-900">{exp.title}</p>
                          <p className="text-slate-600">{exp.company} • {exp.duration}</p>
                          {exp.bullets && (
                            <ul className="mt-2 space-y-1">
                              {exp.bullets.map((bullet: string, j: number) => (
                                <li key={j} className="text-slate-700 text-sm flex items-start gap-2">
                                  <span className="text-blue-500">•</span>
                                  {bullet}
                                </li>
                              ))}
                            </ul>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {resume.education && (
                  <div>
                    <h3 className="font-semibold text-slate-900 mb-2">Education</h3>
                    <div className="space-y-2">
                      {resume.education.map((edu: any, i: number) => (
                        <div key={i}>
                          <p className="font-medium text-slate-900">{edu.degree}</p>
                          <p className="text-slate-600">{edu.institution} • {edu.year}</p>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {resume.achievements && (
                  <div>
                    <h3 className="font-semibold text-slate-900 mb-2">Achievements</h3>
                    <ul className="space-y-1">
                      {resume.achievements.map((ach: string, i: number) => (
                        <li key={i} className="text-slate-700 text-sm flex items-start gap-2">
                          <span className="text-amber-500">★</span>
                          {ach}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>

              <div className="flex gap-4 mt-6">
                <button
                  onClick={handleSaveToStorage}
                  className="flex-1 py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl font-semibold hover:from-blue-700 hover:to-sky-600 transition-all duration-200 shadow-lg shadow-blue-600/25 flex items-center justify-center gap-2"
                >
                  <FileText className="w-5 h-5" />
                  Save Encrypted
                </button>
                <button
                  onClick={() => setStep('input')}
                  className="flex-1 py-3 bg-slate-100 text-slate-700 rounded-xl font-semibold hover:bg-slate-200 transition-all duration-200"
                >
                  Regenerate
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
