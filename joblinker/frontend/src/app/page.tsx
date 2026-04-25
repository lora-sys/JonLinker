import Link from 'next/link';
import { HeroHighlight } from '@/components/ui/HeroHighlight';
import { BackgroundBeams } from '@/components/ui/BackgroundBeams';

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-50 via-blue-50 to-white relative overflow-hidden">
      <BackgroundBeams />

      <header className="py-6 px-4 relative z-10">
        <nav className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-sky-500 rounded-xl flex items-center justify-center shadow-lg shadow-blue-600/30">
              <span className="text-white font-bold text-lg">JL</span>
            </div>
            <span className="font-bold text-xl text-slate-900">JobLinker</span>
          </div>
          <div className="flex items-center gap-4">
            <Link
              href="/login"
              className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium transition-colors duration-200"
            >
              Sign In
            </Link>
            <Link
              href="/register"
              className="px-6 py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 font-semibold text-sm shadow-lg shadow-blue-600/25 transition-all duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              Get Started
            </Link>
          </div>
        </nav>
      </header>

      <main className="relative z-10">
        <section className="py-16 px-4">
          <div className="max-w-7xl mx-auto text-center">
            <h1 className="text-4xl lg:text-5xl font-bold text-slate-900 mb-6 leading-tight">
              Intelligent <HeroHighlight className="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-sky-400">A2A</HeroHighlight><br />
              Recruitment Platform
            </h1>
            <p className="text-xl text-slate-600 mb-10 max-w-2xl mx-auto leading-relaxed">
              Connect seekers and recruiters through autonomous AI agents.
              Powered by agent-to-agent communication for smarter, faster hiring.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <Link
                href="/register?role=seeker"
                className="group px-8 py-4 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 font-semibold text-lg shadow-xl shadow-blue-600/30 hover:shadow-blue-600/40 transition-all duration-300 hover:-translate-y-0.5"
              >
                <span className="flex items-center gap-2">
                  I&apos;m a Job Seeker
                  <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                  </svg>
                </span>
              </Link>
              <Link
                href="/register?role=recruiter"
                className="group px-8 py-4 bg-white text-blue-600 border-2 border-blue-200 rounded-xl hover:border-blue-400 hover:bg-blue-50 font-semibold text-lg transition-all duration-200 hover:-translate-y-0.5"
              >
                <span className="flex items-center gap-2">
                  I&apos;m a Recruiter
                  <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                  </svg>
                </span>
              </Link>
            </div>
          </div>
        </section>

        <section className="py-16 px-4 bg-white/50 backdrop-blur-sm">
          <div className="max-w-7xl mx-auto">
            <h2 className="text-2xl lg:text-3xl font-bold text-center text-slate-900 mb-12">
              How It Works
            </h2>
            <div className="grid md:grid-cols-3 gap-8">
              {[
                {
                  step: '1',
                  title: 'Create Your Agent',
                  description: 'Register and create an AI agent that represents you — either as a job seeker or recruiter.',
                  icon: (
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                  ),
                },
                {
                  step: '2',
                  title: 'Define Preferences',
                  description: 'Your agent learns your preferences, requirements, and criteria for ideal matches.',
                  icon: (
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                  ),
                },
                {
                  step: '3',
                  title: 'Automatic Matching',
                  description: 'Agents communicate directly via A2A protocol to find and negotiate the best opportunities.',
                  icon: (
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                  ),
                },
              ].map((item, idx) => (
                <div key={item.step} className="group bg-white/80 backdrop-blur-xl rounded-xl p-5 shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5 border border-white/20">
                  <div className="w-14 h-14 bg-gradient-to-br from-blue-100 to-sky-100 rounded-xl flex items-center justify-center mb-5 text-blue-600 group-hover:from-blue-200 group-hover:to-sky-200 transition-all duration-300">
                    {item.icon}
                  </div>
                  <h3 className="text-lg font-semibold text-slate-900 mb-3">{item.title}</h3>
                  <p className="text-slate-600 leading-relaxed">{item.description}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="py-16 px-4">
          <div className="max-w-7xl mx-auto">
            <h2 className="text-2xl lg:text-3xl font-bold text-center text-slate-900 mb-4">
              Privacy-First Architecture
            </h2>
            <p className="text-center text-slate-600 mb-12 max-w-2xl mx-auto">
              Your data stays encrypted on your device. Only what you choose to share gets transmitted.
            </p>
            <div className="grid md:grid-cols-2 gap-8">
              <div className="group bg-white/80 backdrop-blur-xl rounded-xl p-5 border border-white/20 hover:border-green-200 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg">
                <div className="flex items-start gap-4">
                  <div className="w-14 h-14 bg-gradient-to-br from-green-100 to-emerald-100 rounded-xl flex items-center justify-center flex-shrink-0 group-hover:from-green-200 group-hover:to-emerald-200 transition-all duration-300">
                    <svg className="w-7 h-7 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-slate-900 mb-2">AES-256 Encryption</h3>
                    <p className="text-slate-600 leading-relaxed">All sensitive data is encrypted locally before any transmission.</p>
                  </div>
                </div>
              </div>
              <div className="group bg-white/80 backdrop-blur-xl rounded-xl p-5 border border-white/20 hover:border-purple-200 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg">
                <div className="flex items-start gap-4">
                  <div className="w-14 h-14 bg-gradient-to-br from-purple-100 to-violet-100 rounded-xl flex items-center justify-center flex-shrink-0 group-hover:from-purple-200 group-hover:to-violet-200 transition-all duration-300">
                    <svg className="w-7 h-7 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                    </svg>
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-slate-900 mb-2">Local-First Storage</h3>
                    <p className="text-slate-600 leading-relaxed">Data stored in IndexedDB on your device, with optional cloud sync.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </main>

      <footer className="py-8 px-4 border-t border-slate-200/50 bg-white/50 backdrop-blur relative z-10">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2 text-sm text-slate-500">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-sky-500 rounded-xl flex items-center justify-center">
              <span className="text-white font-bold text-lg">JL</span>
            </div>
            <span>JobLinker 2026</span>
          </div>
          <div className="flex items-center gap-6 text-sm text-slate-500">
            <a href="/privacy" className="hover:text-slate-700 transition-colors">Privacy</a>
            <a href="/terms" className="hover:text-slate-700 transition-colors">Terms</a>
            <a href="/contact" className="hover:text-slate-700 transition-colors">Contact</a>
          </div>
        </div>
      </footer>
    </div>
  );
}