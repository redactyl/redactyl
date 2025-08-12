package detectors

import (
	"regexp"

	"github.com/redactyl/redactyl/internal/types"
)

var reWeaviateCtx = regexp.MustCompile(`(?i)WEAVIATE_API_KEY|weaviate`)

func WeaviateAPIKey(path string, data []byte) []types.Finding {
	return findWithContext(path, data, reWeaviateCtx, reGenericKey32to64, "weaviate_api_key", types.SevHigh, 0.9)
}
