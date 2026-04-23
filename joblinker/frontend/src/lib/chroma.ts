// Chroma client for vector queries
const CHROMA_URL = process.env.NEXT_PUBLIC_CHROMA_URL || 'http://localhost:8000';

interface ChromaCollection {
  id: string;
  name: string;
  metadata?: Record<string, unknown>;
}

interface ChromaQueryResult {
  ids: string[][];
  distances?: number[][];
  embeddings?: number[][][];
  documents?: string[][];
  metadatas?: Record<string, unknown>[][];
}

interface ChromaAddResult {
  inserted_ids: string[];
}

class ChromaClient {
  private baseUrl: string;

  constructor(baseUrl: string = CHROMA_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(options.headers || {}),
      },
    });

    if (!response.ok) {
      throw new Error(`Chroma request failed: ${response.statusText}`);
    }

    return response.json();
  }

  async listCollections(): Promise<ChromaCollection[]> {
    const result = await this.request<{ collections: ChromaCollection[] }>(
      '/api/v1/collections'
    );
    return result.collections;
  }

  async createCollection(name: string, metadata?: Record<string, unknown>): Promise<ChromaCollection> {
    return this.request<ChromaCollection>('/api/v1/collections', {
      method: 'POST',
      body: JSON.stringify({ name, metadata }),
    });
  }

  async getCollection(name: string): Promise<ChromaCollection> {
    return this.request<ChromaCollection>(`/api/v1/collections/${name}`);
  }

  async deleteCollection(name: string): Promise<void> {
    await this.request(`/api/v1/collections/${name}`, { method: 'DELETE' });
  }

  async add(
    collectionName: string,
    ids: string[],
    embeddings: number[][],
    documents?: string[],
    metadatas?: Record<string, unknown>[]
  ): Promise<ChromaAddResult> {
    return this.request<ChromaAddResult>(`/api/v1/collections/${collectionName}/add`, {
      method: 'POST',
      body: JSON.stringify({
        ids,
        embeddings,
        documents,
        metadatas,
      }),
    });
  }

  async query(
    collectionName: string,
    queryEmbeddings: number[][],
    nResults: number = 10,
    where?: Record<string, unknown>
  ): Promise<ChromaQueryResult> {
    return this.request<ChromaQueryResult>(
      `/api/v1/collections/${collectionName}/query`,
      {
        method: 'POST',
        body: JSON.stringify({
          query_embeddings: queryEmbeddings,
          n_results: nResults,
          where,
        }),
      }
    );
  }

  async get(
    collectionName: string,
    ids: string[],
    include?: ('embeddings' | 'documents' | 'metadatas')[]
  ): Promise<ChromaQueryResult> {
    return this.request<ChromaQueryResult>(
      `/api/v1/collections/${collectionName}/get`,
      {
        method: 'POST',
        body: JSON.stringify({
          ids,
          include,
        }),
      }
    );
  }
}

export const chromaClient = new ChromaClient();
