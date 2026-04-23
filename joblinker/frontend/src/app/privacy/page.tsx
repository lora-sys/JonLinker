'use client';

export default function PrivacyPage() {
  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Privacy Settings</h1>
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 space-y-6">
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-2">Data Export</h2>
          <p className="text-gray-500 mb-4">Download all your data in a portable format</p>
          <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium">
            Export My Data
          </button>
        </div>
        <div className="border-t border-gray-200 pt-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-2">Delete Account</h2>
          <p className="text-gray-500 mb-4">Permanently delete your account and all associated data</p>
          <button className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 font-medium">
            Delete My Account
          </button>
        </div>
      </div>
    </div>
  );
}
