import { FileText } from 'lucide-react'

export default function TermsPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-3xl mx-auto px-4 py-16">
        <div className="flex items-center gap-3 mb-8">
          <FileText className="w-8 h-8 text-blue-600" />
          <h1 className="text-3xl font-bold text-slate-900">Terms of Service</h1>
        </div>
        <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-8 space-y-6 text-slate-600">
          <p>These terms govern your use of the JobLinker platform. By using this service, you agree to these terms.</p>

          <h2 className="text-xl font-semibold text-slate-900">1. Service Description</h2>
          <p>JobLinker is an AI-powered recruitment platform that connects candidates with opportunities through autonomous agent-to-agent negotiations.</p>

          <h2 className="text-xl font-semibold text-slate-900">2. User Responsibilities</h2>
          <p>You are responsible for maintaining the confidentiality of your account credentials and for all activities under your account.</p>

          <h2 className="text-xl font-semibold text-slate-900">3. Data Privacy</h2>
          <p>Your data is handled in accordance with our Privacy Policy. We implement encryption and access controls to protect your information.</p>

          <h2 className="text-xl font-semibold text-slate-900">4. Limitation of Liability</h2>
          <p>JobLinker facilitates connections but does not guarantee employment outcomes. The platform is provided &quot;as is&quot; without warranties.</p>

          <h2 className="text-xl font-semibold text-slate-900">5. Contact</h2>
          <p>For questions about these terms, please contact us through the support channels listed on our Contact page.</p>
        </div>
      </div>
    </div>
  )
}
