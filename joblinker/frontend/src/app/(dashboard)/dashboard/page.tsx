import DashboardClient from '@/shared/ui/DashboardClient'

export default function DashboardPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <DashboardClient />
      </div>
    </div>
  )
}
