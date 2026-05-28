// Package config embeds static configuration assets so the server runs
// independent of the working directory.
package config

import _ "embed"

// TargetingYAML is the embedded targeting configuration.
//
//go:embed targeting.yaml
var TargetingYAML []byte
