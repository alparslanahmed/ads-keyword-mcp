package export

import (
	"encoding/json"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// ToJSON renders a RunExport as indented JSON.
func ToJSON(exp *RunExport) ([]byte, error) {
	b, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		return nil, utils.NewInternalError(err)
	}
	return b, nil
}
