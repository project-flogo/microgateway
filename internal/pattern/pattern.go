//go:generate go-bindata -pkg pattern -o assets.go DefaultHttpPattern.json DefaultChannelPattern.json

package pattern

import (
	"encoding/json"

	"github.com/project-flogo/microgateway/types"
)

// Load loads a pattern
func Load(pattern string) (*types.Microgateway, error) {
	patternJSON, err := Asset(pattern + ".json")
	if err != nil {
		return nil, err
	}
	pDef := &types.Microgateway{}
	err = json.Unmarshal(patternJSON, pDef)
	if err != nil {
		return nil, err
	}
	return pDef, nil
}
