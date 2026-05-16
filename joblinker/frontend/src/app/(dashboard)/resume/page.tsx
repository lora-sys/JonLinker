'use client'

import { FileText, Lock, Sparkles, Upload, X } from 'lucide-react'
import { useCallback, useState } from 'react'

import { initPrivacyStore, privacyStorage } from '@/lib/privacy'

export default function ResumePage() {
  const [isLoading, setIsLoading] = useState(false)
  const [isDragging, setIsDragging] = useState(false)
  const [file, setFile] = useState<File | null>(null)
  const [fileContent, setFileContent] = useState<string | null>(null)
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const [storedResumes, setStoredResumes] = useState<{ id: string, createdAt: string }[]>([])

  const processFile = async (f: File) => {
    setFile(f)
    setError(null)
    const content = await f.text()
    setFileContent(content)
  }

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    const droppedFile = e.dataTransfer.files[0]
    if (droppedFile && (droppedFile.type === 'text/plain' || droppedFile.type === 'application/pdf' || droppedFile.name.endsWith('.md'))) {
      processFile(droppedFile)
    }
    else {
      setError('Please upload a text file (.txt, .md) or PDF')
    }
  }, [])

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0]
    if (selectedFile) {
      processFile(selectedFile)
    }
  }, [])

  const handleUpload = async () => {
    if (!fileContent) {
      setError('Please select a file first')
      return
    }
    if (!password || password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      await initPrivacyStore()
      const id = `resume-${Date.now()}`
      await privacyStorage.storeResume(id, fileContent, password)
      setSuccess(true)
      setFile(null)
      setFileContent(null)
      setPassword('')
      setConfirmPassword('')
      // Refresh stored resumes
      const resumes = await privacyStorage.listResumes()
      setStoredResumes(resumes)
    }
    catch {
      setError('Failed to encrypt and store resume')
    }
    finally {
      setIsLoading(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await privacyStorage.deleteResume(id)
      const resumes = await privacyStorage.listResumes()
      setStoredResumes(resumes)
    }
    catch {
      setError('Failed to delete resume')
    }
  }

  const loadStoredResumes = async () => {
    try {
      await initPrivacyStore()
      const resumes = await privacyStorage.listResumes()
      setStoredResumes(resumes)
    }
    catch {
      // Ignore
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3">
            <FileText className="w-8 h-8 text-blue-600" />
            Resume Storage
          </h1>
          <button
            onClick={loadStoredResumes}
            className="flex items-center gap-2 px-3 py-1.5 text-sm text-blue-600 hover:text-blue-700"
          >
            <Lock className="w-4 h-4" />
            View Stored
          </button>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Upload Section */}
          <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-6">
            <h2 className="text-xl font-semibold text-slate-900 mb-4 flex items-center gap-2">
              <Upload className="w-5 h-5 text-blue-600" />
              Upload Resume
            </h2>

            {error && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                {error}
              </div>
            )}

            {success && (
              <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-green-700 text-sm">
                Resume encrypted and stored securely!
              </div>
            )}

            {/* Drag and Drop Zone */}
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              className={`border-2 border-dashed rounded-xl p-8 text-center transition-colors ${
                isDragging
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-slate-300 hover:border-blue-400 hover:bg-slate-50'
              }`}
            >
              <input
                type="file"
                accept=".txt,.md,.pdf"
                onChange={handleFileSelect}
                className="hidden"
                id="file-upload"
              />
              <label htmlFor="file-upload" className="cursor-pointer">
                <Upload className="w-12 h-12 text-slate-400 mx-auto mb-3" />
                <p className="text-slate-600 font-medium">
                  Drag and drop your resume here, or click to browse
                </p>
                <p className="text-slate-400 text-sm mt-1">
                  Supports .txt, .md, .pdf files
                </p>
              </label>
            </div>

            {file && (
              <div className="mt-4 p-3 bg-slate-50 rounded-lg">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <FileText className="w-4 h-4 text-slate-600" />
                    <span className="text-sm font-medium text-slate-700">{file.name}</span>
                  </div>
                  <button
                    onClick={() => {
                      setFile(null)
                      setFileContent(null)
                    }}
                  >
                    <X className="w-4 h-4 text-slate-400 hover:text-red-500" />
                  </button>
                </div>
              </div>
            )}

            {/* Password Fields */}
            {fileContent && (
              <div className="mt-6 space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1">
                    Encryption Password
                  </label>
                  <input
                    type="password"
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                    placeholder="Min 8 characters"
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1">
                    Confirm Password
                  </label>
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={e => setConfirmPassword(e.target.value)}
                    placeholder="Confirm your password"
                    className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>

                <button
                  onClick={handleUpload}
                  disabled={isLoading}
                  className="w-full py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl font-semibold hover:from-blue-700 hover:to-sky-600 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-lg shadow-blue-600/25"
                >
                  {isLoading
                    ? (
                        <span className="flex items-center justify-center gap-2">
                          <Sparkles className="w-4 h-4 animate-spin" />
                          Encrypting...
                        </span>
                      )
                    : (
                        <span className="flex items-center justify-center gap-2">
                          <Lock className="w-4 h-4" />
                          Encrypt & Store
                        </span>
                      )}
                </button>
              </div>
            )}
          </div>

          {/* Stored Resumes */}
          <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl shadow-lg p-6">
            <h2 className="text-xl font-semibold text-slate-900 mb-4 flex items-center gap-2">
              <Lock className="w-5 h-5 text-green-600" />
              Your Encrypted Resumes
            </h2>

            {storedResumes.length === 0
              ? (
                  <div className="text-center py-8 text-slate-500">
                    <FileText className="w-12 h-12 mx-auto mb-3 text-slate-300" />
                    <p>No resumes stored yet</p>
                    <p className="text-sm">Upload a resume to see it here</p>
                  </div>
                )
              : (
                  <div className="space-y-3">
                    {storedResumes.map(resume => (
                      <div
                        key={resume.id}
                        className="flex items-center justify-between p-3 bg-slate-50 rounded-lg"
                      >
                        <div className="flex items-center gap-3">
                          <FileText className="w-5 h-5 text-slate-600" />
                          <div>
                            <p className="text-sm font-medium text-slate-700">
                              {resume.id}
                            </p>
                            <p className="text-xs text-slate-500">
                              Stored:
                              {' '}
                              {new Date(resume.createdAt).toLocaleDateString()}
                            </p>
                          </div>
                        </div>
                        <button
                          onClick={() => handleDelete(resume.id)}
                          className="p-2 text-slate-400 hover:text-red-500"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
          </div>
        </div>

        {/* Privacy Notice */}
        <div className="mt-6 p-4 bg-amber-50 border border-amber-200 rounded-xl">
          <div className="flex items-start gap-3">
            <Lock className="w-5 h-5 text-amber-600 mt-0.5" />
            <div>
              <h3 className="font-medium text-amber-800">Your Privacy is Protected</h3>
              <p className="text-sm text-amber-700 mt-1">
                All resumes are encrypted locally using AES-256-GCM before being stored in your browser&apos;s IndexedDB.
                Your password never leaves your device. Only you can decrypt and access your resume data.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
