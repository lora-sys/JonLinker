package sessionstore

// EstimateTokenCount provides a rough token count estimate for a set of messages.
// Uses a simple heuristic: ~4 characters per token for English text.
// Returns 0 for empty input.
func EstimateTokenCount(msgs []Message) int {
	total := 0
	for _, m := range msgs {
		total += estimateStringTokens(m.Content)
	}
	return total
}

func estimateStringTokens(s string) int {
	// Rough heuristic: count words + punctuation as tokens
	// ~4 chars per token on average for English
	if len(s) == 0 {
		return 0
	}
	// More accurate for mixed Chinese/English: ~1.5 chars/char for CJK, ~4 for ASCII
	chars := 0
	for _, r := range s {
		if r > 0x4E00 && r < 0x9FFF { // CJK Unified Ideographs
			chars += 2 // ~2 chars per token for Chinese
		} else {
			chars++
		}
	}
	tokens := chars / 4
	if tokens < 1 {
		tokens = 1
	}
	return tokens
}
