export default function PrivacyPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-slate-900 mb-6">Privacy Settings</h1>
        <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-6 space-y-6">
          <div>
            <h2 className="text-lg font-semibold text-slate-900 mb-2">Data Export</h2>
            <p className="text-slate-500 mb-4">Download all your data in a portable format</p>
            <button className="px-6 py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 font-medium shadow-lg shadow-blue-600/25 transition-all duration-200 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2">
              Export My Data
            </button>
          </div>
          <div className="border-t border-slate-200 pt-6">
            <h2 className="text-lg font-semibold text-slate-900 mb-2">Delete Account</h2>
            <p className="text-slate-500 mb-4">Permanently delete your account and all associated data</p>
            <button className="px-6 py-3 bg-gradient-to-r from-red-600 to-rose-500 text-white rounded-xl hover:from-red-700 hover:to-rose-600 font-medium shadow-lg shadow-red-600/25 transition-all duration-200 focus:ring-2 focus:ring-red-500 focus:ring-offset-2">
              Delete My Account
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
