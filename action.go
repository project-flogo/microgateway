package microgateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data"
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

func (mm *MashlingManager) LoadResource(config *resource.Config) (*resource.Resource, error) {

	mashlingDefBytes := config.Data

	var mashlingDefinition *Data
	err := json.Unmarshal(mashlingDefBytes, &mashlingDefinition)
	if err != nil {
		return nil, fmt.Errorf("error marshalling mashling definition resource with id '%s', %s", config.ID, err.Error())
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
	mAction.metadata = &action.Metadata{}
	var actionData *Data
	err := json.Unmarshal(config.Data, &actionData)
	if err != nil {
		return nil, fmt.Errorf("failed to load mashling data: '%s' error '%s'", config.Id, err.Error())
	}
	if actionData.MashlingURI != "" {
		// Load action data from resources
		resData := f.Manager.GetResource(actionData.MashlingURI)
		if resData == nil {
			return nil, fmt.Errorf("failed to load mashling URI data: '%s' error '%s'", config.Id, err.Error())
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

func (m *MashlingAction) Run(context context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {
	payload := make(map[string]interface{})
	for k, v := range inputs {
		payload[k] = v.Value()
	}

	code, mData, err := core.ExecuteMashling(payload, m.configuration, m.dispatch.Routes, m.services)
	output := make(map[string]*data.Attribute)
	codeAttr := data.NewAttribute("code", data.TypeInt, code)
	if err != nil {
		return nil, err
	}
	output["code"] = codeAttr
	dataAttr := data.NewAttribute("data", data.TypeAny, mData)
	if err != nil {
		return nil, err
	}
	output["data"] = dataAttr
	return output, err
}
