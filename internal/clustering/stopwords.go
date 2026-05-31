package clustering

// Stopwords are function words removed before clustering so cluster names key on
// meaningful product terms. Kept intentionally small (function words only) to
// avoid stripping descriptive product vocabulary.
var stopwords = func() map[string]struct{} {
	words := []string{
		// English
		"the", "a", "an", "and", "or", "of", "to", "for", "in", "on", "with",
		"is", "are", "my", "your", "this", "that", "it", "at", "by", "from",
		"as", "be", "you", "i", "we", "me",
		// Turkish
		"ve", "ile", "için", "bir", "bu", "da", "de", "mi", "mı", "mu", "mü",
		"ne", "çok", "daha", "en", "o", "şu", "ki", "ya", "ama", "gibi",
		"ben", "sen", "biz",
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()

// IsStopword reports whether a normalized token is a stopword.
func IsStopword(token string) bool {
	_, ok := stopwords[token]
	return ok
}
