package intent

// Buying modifiers signal commercial value. Multi-word phrases are matched as
// substrings; single tokens are matched on word boundaries. These lists are
// also consumed by the scoring package for the commercial-intent component.
var (
	BuyingModifiersEN = []string{
		"buy", "price", "best", "review", "reviews", "near me", "online",
		"cheap", "discount", "coupon", "alternative", "vs", "for",
	}
	BuyingModifiersTR = []string{
		"satın al", "fiyat", "fiyatları", "en iyi", "yorum", "yorumları",
		"tavsiye", "yakınımda", "online", "ucuz", "indirim", "kupon",
		"alternatif", "karşılaştırma", "için",
	}
)

// Intent classification vocabularies (English + Turkish combined per label).
var (
	commercialTerms = []string{
		"buy", "price", "discount", "coupon",
		"satın al", "fiyat", "indirim", "kupon",
	}
	comparisonTerms = []string{
		"best", "review", "reviews", "vs", "alternative",
		"en iyi", "yorum", "yorumları", "alternatif", "karşılaştırma",
	}
	informationalTerms = []string{
		"how", "what", "why", "guide",
		"nasıl", "nedir", "neden", "rehber",
	}
	problemAwareTerms = []string{
		"pain", "problem", "fix", "solution",
		"ağrı", "sorun", "çözüm", "nasıl geçer",
	}
)
