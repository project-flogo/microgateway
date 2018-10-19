package microgateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/microgateway/internal/pkg/model/v2/action/core"
	"github.com/project-flogo/microgateway/internal/pkg/model/v2/action/pattern"
	"github.com/project-flogo/microgateway/internal/pkg/model/v2/types"
)

const (
	MashlingActionRef = "github.com/project-flogo/microgateway"
)

type MashlingAction struct {
	mashlingURI   string
	metadata      *action.Metadata
	ioMetadata    *metadata.IOMetadata
	dispatch      types.Dispatch
	services      []types.Service
	pattern       string
	configuration map[string]interface{}
}

type Data struct {
	MashlingURI   string                 `json:"mashlingURI"`
	Dispatch      json.RawMessage        `json:"dispatch"`
	Services      json.RawMessage        `json:"services"`
	Pattern       string                 `json:"pattern"`
	Configuration map[string]interface{} `json:"configuration"`
}

type MashlingManager struct {
}

func init() {
	action.Register(&MashlingAction{}, &Factory{})
	resource.RegisterLoader("microgateway", &MashlingManager{})
}

var actionMetadata = action.ToMetadata(&Settings{}, &Input{}, &Output{})

func (mm *MashlingManager) LoadResource(config *resource.Config) (*resource.Resource, error) {

	mashlingDefBytes := config.Data

	var mashlingDefinition *Data
	err := json.Unmarshal(mashlingDefBytes, &mashlingDefinition)
	if err != nil {
		return nil, fmt.Errorf("error marshalling microgateway definition resource with id '%s', %s", config.ID, err.Error())
	}

	return resource.New("microgateway", mashlingDefinition), nil
}

type Factory struct {
	*resource.Manager
}

func (f *Factory) Initialize(ctx action.InitContext) error {
	f.Manager = ctx.ResourceManager()
	return nil
}

func (f *Factory) New(config *action.Config) (action.Action, error) {
	mAction := &MashlingAction{}
	mAction.metadata = actionMetadata

	var actionData *Data
	err := json.Unmarshal(config.Data, &config.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to load microgateway data: '%s' error '%s'", config.Id, err.Error())
	}

	s := &Settings{}
	err = metadata.MapToStruct(config.Settings, s, true)
	if err != nil {
		return nil, err
	}

	if s.URI != "" {
		// Load action data from resources
		resData := f.Manager.GetResource(s.URI)
		if resData == nil {
			return nil, fmt.Errorf("failed to load microgateway URI data: '%s' error '%s'", config.Id, err.Error())
		}
		actionData = resData.Object().(*Data)
	}
	// Extract configuration
	mAction.configuration = actionData.Configuration
	// Extract pattern
	mAction.pattern = actionData.Pattern
	if mAction.pattern == "" {
		// Parse routes
		var dispatch types.Dispatch
		err = json.Unmarshal(actionData.Dispatch, &dispatch)
		if err != nil {
			return nil, err
		}
		// Parse services
		var services []types.Service
		err = json.Unmarshal(actionData.Services, &services)
		if err != nil {
			return nil, err
		}
		mAction.dispatch = dispatch
		mAction.services = services
	} else {
		pDef, err := pattern.Load(mAction.pattern)
		if err != nil {
			return nil, err
		}
		mAction.dispatch = pDef.Dispatch
		mAction.services = pDef.Services
	}

	return mAction, nil
}

func (m *MashlingAction) Metadata() *action.Metadata {
	return m.metadata
}

func (m *MashlingAction) IOMetadata() *metadata.IOMetadata {
	return m.ioMetadata
}

func (m *MashlingAction) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	code, mData, err := core.ExecuteMashling(input, m.configuration, m.dispatch.Routes, m.services)
	output := make(map[string]interface{}, 8)
	output["code"] = code
	output["data"] = mData

	return output, err
}
