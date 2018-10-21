package microgateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data/expression"
	_ "github.com/project-flogo/core/data/expression/script"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/microgateway/internal/pkg/model/v2/action/core"
	"github.com/project-flogo/microgateway/internal/pkg/model/v2/action/pattern"
	"github.com/project-flogo/microgateway/internal/pkg/model/v2/types"
)

type Action struct {
	mashlingURI   string
	metadata      *action.Metadata
	ioMetadata    *metadata.IOMetadata
	dispatch      types.Dispatch
	services      []types.Service
	serviceCache  map[string]*types.Service
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

type Manager struct {
}

func init() {
	action.Register(&Action{}, &Factory{})
	resource.RegisterLoader("microgateway", &Manager{})
}

var actionMetadata = action.ToMetadata(&Settings{}, &Input{}, &Output{})

func (m *Manager) LoadResource(config *resource.Config) (*resource.Resource, error) {
	data := config.Data

	var definition *Data
	err := json.Unmarshal(data, &definition)
	if err != nil {
		return nil, fmt.Errorf("error marshalling microgateway definition resource with id '%s', %s", config.ID, err.Error())
	}

	return resource.New("microgateway", definition), nil
}

type Factory struct {
	*resource.Manager
}

type initContext struct {
	settings map[string]interface{}
}

func (i *initContext) Settings() map[string]interface{} {
	return i.settings
}

func (i *initContext) MapperFactory() mapper.Factory {
	return nil
}

func (f *Factory) Initialize(ctx action.InitContext) error {
	f.Manager = ctx.ResourceManager()
	return nil
}

func (f *Factory) New(config *action.Config) (action.Action, error) {
	act := &Action{}
	act.metadata = actionMetadata

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
	act.configuration = actionData.Configuration
	// Extract pattern
	act.pattern = actionData.Pattern
	if act.pattern == "" {
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
		act.dispatch = dispatch
		act.services = services
	} else {
		pDef, err := pattern.Load(act.pattern)
		if err != nil {
			return nil, err
		}
		act.dispatch = pDef.Dispatch
		act.services = pDef.Services
	}

	expressionFactory := expression.NewFactory(resolve.GetBasicResolver())
	getExpression := func(value interface{}) (expression.Expr, error) {
		if stringValue, ok := value.(string); ok && len(stringValue) > 0 && stringValue[0] == '=' {
			expr, err := expressionFactory.NewExpr(stringValue[1:])
			if err != nil {
				return nil, err
			}
			return expr, nil
		}
		return expression.NewLiteralExpr(value), nil
	}
	routes := act.dispatch.Routes
	for i := range routes {
		if condition := routes[i].Condition; condition != "" {
			expr, err := expressionFactory.NewExpr(condition)
			if err != nil {
				log.Infof("condition parsing error: %s", condition)
				return nil, err
			}
			routes[i].Expression = expr
		}
		steps := routes[i].Steps
		for j := range steps {
			if condition := steps[j].Condition; condition != "" {
				expr, err := expressionFactory.NewExpr(condition)
				if err != nil {
					log.Infof("condition parsing error: %s", condition)
					return nil, err
				}
				steps[j].Expression = expr
			}
			input := steps[j].Input
			inputExpression := make(map[string]expression.Expr, len(input))
			for key, value := range input {
				inputExpression[key], err = getExpression(value)
				if err != nil {
					return nil, err
				}
			}
			steps[j].InputExpression = inputExpression
		}
		responses := routes[i].Responses
		for j := range responses {
			if condition := responses[j].Condition; condition != "" {
				expr, err := expressionFactory.NewExpr(condition)
				if err != nil {
					log.Infof("condition parsing error: %s", condition)
					return nil, err
				}
				responses[j].Expression = expr
			}
			responses[j].Output.CodeExpression, err = getExpression(responses[j].Output.Code)
			if err != nil {
				return nil, err
			}
			data := responses[j].Output.Data
			if hashMap, ok := data.(map[string]interface{}); ok {
				dataExpressions := make(map[string]expression.Expr, len(hashMap))
				for key, value := range hashMap {
					dataExpressions[key], err = getExpression(value)
					if err != nil {
						return nil, err
					}
				}
				responses[j].Output.DataExpressions = dataExpressions
			} else {
				responses[j].Output.DataExpression, err = getExpression(data)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	serviceCache := make(map[string]*types.Service, len(act.services))
	for i := range act.services {
		name := act.services[i].Name
		if _, ok := serviceCache[name]; ok {
			return nil, fmt.Errorf("duplicate service name: %s", name)
		}
		ref := act.services[i].Ref
		if factory := activity.GetFactory(ref); factory != nil {
			actvt, err := factory(&initContext{settings: act.services[i].Settings})
			if err != nil {
				return nil, err
			}
			act.services[i].Activity = actvt
			serviceCache[name] = &act.services[i]
			continue
		}
		actvt := activity.Get(ref)
		if actvt == nil {
			return nil, fmt.Errorf("can't find activity %s", ref)
		}
		act.services[i].Activity = actvt
		serviceCache[name] = &act.services[i]
	}
	act.serviceCache = serviceCache

	return act, nil
}

func (m *Action) Metadata() *action.Metadata {
	return m.metadata
}

func (m *Action) IOMetadata() *metadata.IOMetadata {
	return m.ioMetadata
}

func (m *Action) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	code, mData, err := core.Execute(input, m.configuration, m.dispatch.Routes, m.serviceCache)
	output := make(map[string]interface{}, 8)
	output["code"] = code
	output["data"] = mData

	return output, err
}
