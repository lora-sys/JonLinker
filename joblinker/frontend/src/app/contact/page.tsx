import { Mail, MessageSquare } from 'lucide-react'

export default function ContactPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-3xl mx-auto px-4 py-16">
        <div className="flex items-center gap-3 mb-8">
          <MessageSquare className="w-8 h-8 text-blue-600" />
          <h1 className="text-3xl font-bold text-slate-900">Contact Us</h1>
        </div>
        <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-8 space-y-6">
          <div className="flex items-start gap-4 p-4 bg-blue-50 rounded-xl">
            <Mail className="w-6 h-6 text-blue-600 mt-0.5" />
            <div>
              <h2 className="font-semibold text-slate-900">Email</h2>
              <p className="text-slate-600">support@joblinker.dev</p>
            </div>
          </div>
          <div className="flex items-start gap-4 p-4 bg-green-50 rounded-xl">
            <MessageSquare className="w-6 h-6 text-green-600 mt-0.5" />
            <div>
              <h2 className="font-semibold text-slate-900">GitHub</h2>
              <p className="text-slate-600">github.com/joblinker/joblinker</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
