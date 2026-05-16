// IndexedDB integration for local-first storage
import type { Agent, Interview, Job, Match, Message, Offer, Resume } from '@/types'

const DB_NAME = 'joblinker'
const DB_VERSION = 1

interface DBSchema {
  agents: Agent
  jobs: Job
  resumes: Resume
  matches: Match
  messages: Message
  interviews: Interview
  offers: Offer
}

let dbInstance: IDBDatabase | null = null

export async function getDB(): Promise<IDBDatabase> {
  if (dbInstance)
    return dbInstance

  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)

    request.onerror = () => reject(request.error)
    request.onsuccess = () => {
      dbInstance = request.result
      resolve(request.result)
    }

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result

      if (!db.objectStoreNames.contains('agents')) {
        db.createObjectStore('agents', { keyPath: 'id' })
      }
      if (!db.objectStoreNames.contains('jobs')) {
        db.createObjectStore('jobs', { keyPath: 'id' })
      }
      if (!db.objectStoreNames.contains('resumes')) {
        db.createObjectStore('resumes', { keyPath: 'id' })
      }
      if (!db.objectStoreNames.contains('matches')) {
        db.createObjectStore('matches', { keyPath: 'id' })
      }
      if (!db.objectStoreNames.contains('messages')) {
        const msgStore = db.createObjectStore('messages', { keyPath: 'id' })
        msgStore.createIndex('match_id', 'match_id', { unique: false })
      }
      if (!db.objectStoreNames.contains('interviews')) {
        db.createObjectStore('interviews', { keyPath: 'id' })
      }
      if (!db.objectStoreNames.contains('offers')) {
        db.createObjectStore('offers', { keyPath: 'id' })
      }
    }
  })
}

export async function getAll<T extends keyof DBSchema>(
  storeName: T,
): Promise<DBSchema[T][]> {
  const db = await getDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, 'readonly')
    const store = tx.objectStore(storeName)
    const request = store.getAll()
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

export async function get<T extends keyof DBSchema>(
  storeName: T,
  id: string,
): Promise<DBSchema[T] | undefined> {
  const db = await getDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, 'readonly')
    const store = tx.objectStore(storeName)
    const request = store.get(id)
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

export async function put<T extends keyof DBSchema>(
  storeName: T,
  item: DBSchema[T],
): Promise<void> {
  const db = await getDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, 'readwrite')
    const store = tx.objectStore(storeName)
    const request = store.put(item)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

export async function remove<T extends keyof DBSchema>(
  storeName: T,
  id: string,
): Promise<void> {
  const db = await getDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, 'readwrite')
    const store = tx.objectStore(storeName)
    const request = store.delete(id)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

export async function clearStore<T extends keyof DBSchema>(
  storeName: T,
): Promise<void> {
  const db = await getDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, 'readwrite')
    const store = tx.objectStore(storeName)
    const request = store.clear()
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

export async function queryByIndex<T extends keyof DBSchema>(
  storeName: T,
  indexName: string,
  value: IDBValidKey,
): Promise<DBSchema[T][]> {
  const db = await getDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, 'readonly')
    const store = tx.objectStore(storeName)
    const index = store.index(indexName)
    const request = index.getAll(value)
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}
