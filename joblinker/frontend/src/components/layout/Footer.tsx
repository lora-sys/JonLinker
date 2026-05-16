'use client'

import Image from 'next/image'

export function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-white border-t border-gray-200 py-6">
      <div className="px-6 flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <div className="relative w-6 h-6">
            <Image
              src="/logo.png"
              alt="JobLinker"
              fill
              className="object-contain"
            />
          </div>
          <span>
            JobLinker
            {currentYear}
          </span>
        </div>
        <div className="flex items-center gap-6 text-sm text-gray-500">
          <a href="/privacy" className="hover:text-gray-700 transition-colors">
            Privacy Policy
          </a>
          <a href="/terms" className="hover:text-gray-700 transition-colors">
            Terms of Service
          </a>
          <a href="/contact" className="hover:text-gray-700 transition-colors">
            Contact
          </a>
        </div>
      </div>
    </footer>
  )
}
