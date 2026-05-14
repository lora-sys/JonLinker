// Protobuf parsing utilities for Gateway client
// Handles decoding of binary Protobuf responses from backend

/**
 * Decode a Protobuf binary response
 * Falls back to raw ArrayBuffer if decoding fails
 */
export async function parseProtobufResponse<T>(
  buffer: ArrayBuffer,
  messageType?: string
): Promise<T> {
  // For now, return raw binary data cast to T
  // In production, would use generated stubs from backend/pkg/proto/
  // const protobuf = await import('protobufjs');
  const binaryData = new Uint8Array(buffer);
  return binaryData as unknown as T;
}

/**
 * Detect if a response is Protobuf-encoded
 */
export function isProtobufResponse(contentType: string | null): boolean {
  if (!contentType) return false;
  return contentType.includes('x-protobuf') || contentType.includes('protobuf');
}

/**
 * Convert ArrayBuffer to readable format for debugging
 */
export function arrayBufferToHex(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  return Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join(' ');
}