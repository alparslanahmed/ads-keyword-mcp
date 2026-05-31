package googleads

import (
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alparslanahmed/keyword-market-mcp/config"
	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// Targeting is a resolved country/language pair with its Google Ads constants.
type Targeting struct {
	Country           string `json:"country"`
	Language          string `json:"language"`
	GeoTargetConstant string `json:"geoTargetConstant"`
	LanguageConstant  string `json:"languageConstant"`
}

// CountryInfo / LanguageInfo are listing entries for the config resource.
type CountryInfo struct {
	Name              string `json:"name"`
	GeoTargetConstant string `json:"geoTargetConstant"`
}
type LanguageInfo struct {
	Name             string `json:"name"`
	LanguageConstant string `json:"languageConstant"`
}

// TargetingRegistry resolves human-readable names to constants.
type TargetingRegistry struct {
	countries map[string]CountryInfo // keyed by lowercased name
	languages map[string]LanguageInfo
}

type targetingYAML struct {
	Countries map[string]struct {
		GeoTargetConstant string `yaml:"geoTargetConstant"`
	} `yaml:"countries"`
	Languages map[string]struct {
		LanguageConstant string `yaml:"languageConstant"`
	} `yaml:"languages"`
}

// LoadTargetingRegistry parses the embedded targeting configuration.
func LoadTargetingRegistry() (*TargetingRegistry, error) {
	var parsed targetingYAML
	if err := yaml.Unmarshal(config.TargetingYAML, &parsed); err != nil {
		return nil, utils.NewInternalError(err)
	}
	r := &TargetingRegistry{
		countries: make(map[string]CountryInfo),
		languages: make(map[string]LanguageInfo),
	}
	for name, v := range parsed.Countries {
		r.countries[strings.ToLower(name)] = CountryInfo{Name: name, GeoTargetConstant: v.GeoTargetConstant}
	}
	for name, v := range parsed.Languages {
		r.languages[strings.ToLower(name)] = LanguageInfo{Name: name, LanguageConstant: v.LanguageConstant}
	}
	return r, nil
}

// Resolve maps a country/language pair (case-insensitive) to constants.
func (r *TargetingRegistry) Resolve(country, language string) (Targeting, error) {
	c, ok := r.countries[strings.ToLower(strings.TrimSpace(country))]
	if !ok {
		return Targeting{}, utils.NewInvalidTargetingError("Unsupported country: " + country)
	}
	l, ok := r.languages[strings.ToLower(strings.TrimSpace(language))]
	if !ok {
		return Targeting{}, utils.NewInvalidTargetingError("Unsupported language: " + language)
	}
	return Targeting{
		Country:           c.Name,
		Language:          l.Name,
		GeoTargetConstant: c.GeoTargetConstant,
		LanguageConstant:  l.LanguageConstant,
	}, nil
}

// Countries returns supported countries sorted by name.
func (r *TargetingRegistry) Countries() []CountryInfo {
	out := make([]CountryInfo, 0, len(r.countries))
	for _, c := range r.countries {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Languages returns supported languages sorted by name.
func (r *TargetingRegistry) Languages() []LanguageInfo {
	out := make([]LanguageInfo, 0, len(r.languages))
	for _, l := range r.languages {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
