// Local vector generation using simple TF-IDF approach
// For production, consider using WebAssembly-based embedding models

interface TokenFrequency {
  [token: string]: number;
}

export function tokenize(text: string): string[] {
  return text
    .toLowerCase()
    .replace(/[^\w\s]/g, ' ')
    .split(/\s+/)
    .filter((token) => token.length > 2);
}

export function computeTF(tokens: string[]): TokenFrequency {
  const tf: TokenFrequency = {};
  for (const token of tokens) {
    tf[token] = (tf[token] || 0) + 1;
  }
  const len = tokens.length;
  if (len === 0) return tf;
  for (const token in tf) {
    const val = tf[token];
    if (val !== undefined) {
      tf[token] = val / len;
    }
  }
  return tf;
}

export function computeIDF(
  documents: string[][]
): TokenFrequency {
  const idf: TokenFrequency = {};
  const numDocs = documents.length;

  const allTokens = new Set(documents.flat());
  for (const token of allTokens) {
    let docCount = 0;
    for (const doc of documents) {
      if (doc.includes(token)) docCount++;
    }
    idf[token] = Math.log(numDocs / (1 + docCount));
  }

  return idf;
}

export function computeTFIDF(
  tf: TokenFrequency,
  idf: TokenFrequency
): number[] {
  const tokens = Object.keys({ ...tf, ...idf });
  return tokens.map((token) => (tf[token] || 0) * (idf[token] || 0));
}

export function cosineSimilarity(a: number[], b: number[]): number {
  if (a.length !== b.length) return 0;

  let dotProduct = 0;
  let normA = 0;
  let normB = 0;

  for (let i = 0; i < a.length; i++) {
    const aVal = a[i];
    const bVal = b[i];
    if (aVal === undefined || bVal === undefined) continue;
    dotProduct += aVal * bVal;
    normA += aVal * aVal;
    normB += bVal * bVal;
  }

  const denominator = Math.sqrt(normA) * Math.sqrt(normB);
  return denominator === 0 ? 0 : dotProduct / denominator;
}

export function generateEmbedding(text: string): number[] {
  const tokens = tokenize(text);
  const tf = computeTF(tokens);
  return Object.values(tf);
}

export function generateEmbeddingBatch(texts: string[]): number[][] {
  const tokenizedTexts = texts.map(tokenize);
  const allTokens = [...new Set(tokenizedTexts.flat())];
  const idf = computeIDF(tokenizedTexts);

  return tokenizedTexts.map((tokens) => {
    const tf = computeTF(tokens);
    return allTokens.map((token) => (tf[token] || 0) * (idf[token] || 0));
  });
}
