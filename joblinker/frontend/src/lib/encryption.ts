// AES-256-GCM encryption utilities for local-first privacy
const ALGORITHM = 'AES-GCM'
const KEY_LENGTH = 256
const IV_LENGTH = 12
const TAG_LENGTH = 128

async function deriveKey(password: string, salt: Uint8Array): Promise<CryptoKey> {
  const encoder = new TextEncoder()
  const passwordBuffer = encoder.encode(password)

  const baseKey = await crypto.subtle.importKey(
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
    baseKey,
    { name: ALGORITHM, length: KEY_LENGTH },
    false,
    ['encrypt', 'decrypt'],
  )
}

export async function encrypt(
  plaintext: string,
  password: string,
): Promise<{ ciphertext: string, salt: string, iv: string }> {
  const encoder = new TextEncoder()
  const salt = crypto.getRandomValues(new Uint8Array(16))
  const iv = crypto.getRandomValues(new Uint8Array(IV_LENGTH))

  const key = await deriveKey(password, salt)

  const ciphertext = await crypto.subtle.encrypt(
    { name: ALGORITHM, iv: iv.buffer as ArrayBuffer, tagLength: TAG_LENGTH },
    key,
    encoder.encode(plaintext),
  )

  return {
    ciphertext: bufferToBase64(new Uint8Array(ciphertext)),
    salt: bufferToBase64(salt),
    iv: bufferToBase64(new Uint8Array(iv)),
  }
}

export async function decrypt(
  ciphertext: string,
  password: string,
  salt: string,
  iv: string,
): Promise<string> {
  const decoder = new TextDecoder()

  const key = await deriveKey(password, base64ToBuffer(salt))

  const plaintext = await crypto.subtle.decrypt(
    { name: ALGORITHM, iv: base64ToBuffer(iv).buffer as ArrayBuffer, tagLength: TAG_LENGTH },
    key,
    base64ToBuffer(ciphertext).buffer as ArrayBuffer,
  )

  return decoder.decode(plaintext)
}

function bufferToBase64(buffer: Uint8Array): string {
  let binary = ''
  buffer.forEach(byte => (binary += String.fromCharCode(byte)))
  return btoa(binary)
}

function base64ToBuffer(base64: string): Uint8Array {
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

export function generatePassword(length = 32): string {
  const charset
    = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  const values = crypto.getRandomValues(new Uint8Array(length))
  return Array.from(values)
    .map(v => charset[v % charset.length])
    .join('')
}
