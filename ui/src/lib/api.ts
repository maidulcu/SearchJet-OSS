const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface SearchResult {
  hits: Array<{
    id: string;
    title: string;
    body: string;
    lang: string;
    emirate?: string;
    source: string;
  }>;
  total: number;
  page: number;
  limit: number;
  processing_ms: number;
  query_id: string;
}

export interface Document {
  id?: string;
  title: string;
  body: string;
  lang: string;
  emirate?: string;
  source: string;
}

export async function search(query: string, options?: {
  lang?: string;
  emirate?: string;
  lat?: number;
  lng?: number;
  page?: number;
  limit?: number;
}): Promise<SearchResult> {
  const params = new URLSearchParams({ q: query });
  if (options?.lang) params.set('lang', options.lang);
  if (options?.emirate) params.set('emirate', options.emirate);
  if (options?.lat) params.set('lat', String(options.lat));
  if (options?.lng) params.set('lng', String(options.lng));
  if (options?.page) params.set('page', String(options.page));
  if (options?.limit) params.set('limit', String(options.limit));

  const res = await fetch(`${API_BASE}/v1/search?${params}`);
  if (!res.ok) throw new Error('Search failed');
  return res.json();
}

export async function indexDocuments(documents: Document[]): Promise<{ indexed: number }> {
  const res = await fetch(`${API_BASE}/v1/index`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ documents }),
  });
  if (!res.ok) throw new Error('Index failed');
  return res.json();
}

export async function deleteDocument(id: string): Promise<{ deleted: string }> {
  const res = await fetch(`${API_BASE}/v1/index/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Delete failed');
  return res.json();
}

export async function healthCheck(): Promise<{ status: string }> {
  const res = await fetch(`${API_BASE}/v1/health`);
  if (!res.ok) throw new Error('Health check failed');
  return res.json();
}
