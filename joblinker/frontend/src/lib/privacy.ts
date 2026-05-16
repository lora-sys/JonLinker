'use client'

import type { DBSchema, IDBPDatabase } from 'idb'

import { openDB } from 'idb'

interface ResumeRecord {
  id: string
  encryptedData: string
  iv: string
  salt: string
  createdAt: string
  updatedAt: string
}

interface JobLinkerPrivacyDB extends DBSchema {
  resumes: {
    key: string
    value: ResumeRecord
    indexes: { 'by-created': string }
  }
}

const DB_NAME = 'joblinker-privacy'
const DB_VERSION = 1

export class PrivacyStorage {
  private db: IDBPDatabase<JobLinkerPrivacyDB> | null = null
  private encryptionKey: CryptoKey | null = null

  async init(): Promise<void> {
    this.db = await openDB<JobLinkerPrivacyDB>(DB_NAME, DB_VERSION, {
      upgrade(db) {
        if (!db.objectStoreNames.contains('resumes')) {
          const store = db.createObjectStore('resumes', { keyPath: 'id' })
          store.createIndex('by-created', 'createdAt')
        }
      },
    })
  }

  async deriveKey(password: string, salt: Uint8Array): Promise<CryptoKey> {
    const encoder = new TextEncoder()
    const passwordBuffer = encoder.encode(password)

    const keyMaterial = await crypto.subtle.importKey(
      'raw',
      passwordBuffer,
      'PBKDF2',
      false,
      ['deriveKey'],
    )

    return crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt: salt.buffer as ArrayBuffer,
        iterations: 100000,
        hash: 'SHA-256',
      },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt'],
    )
  }

  async encryptData(data: string, password: string): Promise<{ encrypted: string, iv: string, salt: string }> {
    const salt = crypto.getRandomValues(new Uint8Array(16))
    const iv = crypto.getRandomValues(new Uint8Array(12))
    const key = await this.deriveKey(password, salt)

    const encoder = new TextEncoder()
    const dataBuffer = encoder.encode(data)

    const encryptedBuffer = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv: iv.buffer as ArrayBuffer },
      key,
      dataBuffer,
    )

    return {
      encrypted: btoa(String.fromCharCode(...new Uint8Array(encryptedBuffer))),
      iv: btoa(String.fromCharCode(...iv)),
      salt: btoa(String.fromCharCode(...salt)),
    }
  }

  async decryptData(encrypted: string, iv: string, salt: string, password: string): Promise<string> {
    const saltArray = Uint8Array.from(atob(salt), c => c.charCodeAt(0))
    const ivArray = Uint8Array.from(atob(iv), c => c.charCodeAt(0))
    const encryptedArray = Uint8Array.from(atob(encrypted), c => c.charCodeAt(0))

    const key = await this.deriveKey(password, saltArray)

    const decryptedBuffer = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv: ivArray.buffer as ArrayBuffer },
      key,
      encryptedArray,
    )

    const decoder = new TextDecoder()
    return decoder.decode(decryptedBuffer)
  }

  async storeResume(id: string, data: string, password: string): Promise<void> {
    if (!this.db)
      await this.init()

    const { encrypted, iv, salt } = await this.encryptData(data, password)
    const now = new Date().toISOString()

    await this.db!.put('resumes', {
      id,
      encryptedData: encrypted,
      iv,
      salt,
      createdAt: now,
      updatedAt: now,
    })
  }

  async getResume(id: string, password: string): Promise<string | null> {
    if (!this.db)
      await this.init()

    const record = await this.db!.get('resumes', id)
    if (!record)
      return null

    return this.decryptData(record.encryptedData, record.iv, record.salt, password)
  }

  async deleteResume(id: string): Promise<void> {
    if (!this.db)
      await this.init()
    await this.db!.delete('resumes', id)
  }

  async listResumes(): Promise<{ id: string, createdAt: string }[]> {
    if (!this.db)
      await this.init()
    return this.db!.getAllFromIndex('resumes', 'by-created')
  }
}

export const privacyStorage = new PrivacyStorage()

export async function initPrivacyStore(): Promise<void> {
  await privacyStorage.init()
}
