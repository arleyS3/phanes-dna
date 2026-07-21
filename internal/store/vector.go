package store

import (
	"context"
	"math"
	"sort"

	"github.com/arley/phanes-dna/internal/dna"
)

// SearchSimilar performs brute-force cosine similarity search over all chunks
// that have a non-nil embedding BLOB and returns the top-K results ordered by
// descending similarity. Cosine similarity is computed as:
//
//	similarity = dot(a,b) / (|a| * |b|)
//
// Distance = 1 - similarity.
func (s *Store) SearchSimilar(ctx context.Context, embedding []float32, topK int) ([]*dna.Chunk, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, chunk_idx, content, embedding FROM chunks WHERE embedding IS NOT NULL ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		chunk      *dna.Chunk
		similarity float64
	}
	var candidates []scored

	for rows.Next() {
		var c dna.Chunk
		var id int64
		var blob []byte
		if err := rows.Scan(&id, &c.StartLine, &c.Content, &blob); err != nil {
			return nil, err
		}
		if blob == nil {
			continue
		}
		emb := bytesToFloats(blob)
		sim := cosineSimilarity(embedding, emb)
		candidates = append(candidates, scored{chunk: &c, similarity: sim})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].similarity > candidates[j].similarity
	})
	if topK > 0 && len(candidates) > topK {
		candidates = candidates[:topK]
	}

	result := make([]*dna.Chunk, len(candidates))
	for i, sc := range candidates {
		sc.chunk.ID = "" // ID is store-internal; leave empty for external use
		result[i] = sc.chunk
	}
	return result, nil
}

// cosineSimilarity returns the cosine similarity between two float32 vectors.
// Returns 0 if either vector is zero-length.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		va := float64(a[i])
		vb := float64(b[i])
		dot += va * vb
		normA += va * va
		normB += vb * vb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
