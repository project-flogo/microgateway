package microgateway

import (
	"context"
	"encoding/json"
	"fmt"

	_ "github.com/project-flogo/contrib/function"
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/app/resource"
	"github.com/project-flogo/core/data/expression"
	_ "github.com/project-flogo/core/data/expression/script"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support/logger"
	"github.com/project-flogo/microgateway/internal/core"
	"github.com/project-flogo/microgateway/internal/pattern"
	"github.com/project-flogo/microgateway/internal/types"
)

var log = logger.GetLogger("microgateway")

type Action struct {
	id           string
	ioMetadata   *metadata.IOMetadata
	settings     Settings
	microgateway *core.Microgateway
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

	var definition *types.Microgateway
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
	act := Action{
		id: config.Id,
	}
	if act.id == "" {
		act.id = config.Ref
	}

	if len(config.Data) > 0 {
		err := json.Unmarshal(config.Data, &config.Settings)
		if err != nil {
			return nil, err
		}
	}

	err := metadata.MapToStruct(config.Settings, &act.settings, true)
	if err != nil {
		return nil, err
	}

	// Load action data from resources
	resData := f.Manager.GetResource(act.settings.URI)
	if resData == nil {
		return nil, fmt.Errorf("failed to load microgateway URI data: '%s'", config.Id)
	}
	actionData := resData.Object().(*types.Microgateway)

	if actionData.Pattern != "" {
		definition, err := pattern.Load(actionData.Pattern)
		if err != nil {
			return nil, err
		}
		definition.Name = actionData.Name
		definition.Pattern = actionData.Pattern
		definition.Async = actionData.Async
		definition.Configuration = actionData.Configuration
		actionData = definition
	}

	services := make(map[string]*core.Service, len(actionData.Services))
	for i := range actionData.Services {
		name := actionData.Services[i].Name
		if _, ok := services[name]; ok {
			return nil, fmt.Errorf("duplicate service name: %s", name)
		}
		ref := actionData.Services[i].Ref
		if factory := activity.GetFactory(ref); factory != nil {
			actvt, err := factory(&initContext{settings: actionData.Services[i].Settings})
			if err != nil {
				return nil, err
			}
			services[name] = &core.Service{
				Name:     name,
				Settings: actionData.Services[i].Settings,
				Activity: actvt,
			}
			continue
		}
		actvt := activity.Get(ref)
		if actvt == nil {
			return nil, fmt.Errorf("can't find activity %s", ref)
		}
		services[name] = &core.Service{
			Name:     name,
			Settings: actionData.Services[i].Settings,
			Activity: actvt,
		}
	}

	expressionFactory := expression.NewFactory(resolve.GetBasicResolver())
	getExpression := func(value interface{}) (*types.Expr, error) {
		if stringValue, ok := value.(string); ok && len(stringValue) > 0 && stringValue[0] == '=' {
			expr, err := expressionFactory.NewExpr(stringValue[1:])
			if err != nil {
				return nil, err
			}
			return types.NewExpr(stringValue, expr), nil
		}
		return types.NewExpr(fmt.Sprintf("%v", value), expression.NewLiteralExpr(value)), nil
	}

	steps, responses := actionData.Steps, actionData.Responses
	microgateway := core.Microgateway{
		Name:          actionData.Name,
		Async:         actionData.Async,
		Steps:         make([]core.Step, len(steps)),
		Responses:     make([]core.Response, len(responses)),
		Configuration: actionData.Configuration,
	}
	for j := range steps {
		if condition := steps[j].Condition; condition != "" {
			expr, err := expressionFactory.NewExpr(condition)
			if err != nil {
				log.Infof("condition parsing error: %s", condition)
				return nil, err
			}
			microgateway.Steps[j].Condition = types.NewExpr(condition, expr)
		}

		service := services[steps[j].Service]
		if service == nil {
			return nil, fmt.Errorf("service not found: %s", steps[j].Service)
		}
		microgateway.Steps[j].Service = service

		input := steps[j].Input
		inputExpression := make(map[string]*types.Expr, len(input))
		for key, value := range input {
			inputExpression[key], err = getExpression(value)
			if err != nil {
				return nil, err
			}
		}
		microgateway.Steps[j].Input = inputExpression
	}

	for j := range responses {
		if condition := responses[j].Condition; condition != "" {
			expr, err := expressionFactory.NewExpr(condition)
			if err != nil {
				log.Infof("condition parsing error: %s", condition)
				return nil, err
			}
			microgateway.Responses[j].Condition = types.NewExpr(condition, expr)
		}

		microgateway.Responses[j].Error = responses[j].Error

		microgateway.Responses[j].Output.Code, err = getExpression(responses[j].Output.Code)
		if err != nil {
			return nil, err
		}

		data := responses[j].Output.Data
		if hashMap, ok := data.(map[string]interface{}); ok {
			dataExpressions := make(map[string]*types.Expr, len(hashMap))
			for key, value := range hashMap {
				dataExpressions[key], err = getExpression(value)
				if err != nil {
					return nil, err
				}
			}
			microgateway.Responses[j].Output.Datum = dataExpressions
		} else {
			microgateway.Responses[j].Output.Data, err = getExpression(data)
			if err != nil {
				return nil, err
			}
		}
	}

	act.microgateway = &microgateway

	return &act, nil
}

func (a *Action) Metadata() *action.Metadata {
	return actionMetadata
}

func (a *Action) IOMetadata() *metadata.IOMetadata {
	return a.ioMetadata
}

func (a *Action) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	code, mData, err := core.Execute(a.id, input, a.microgateway)
	output := make(map[string]interface{}, 8)
	output["code"] = code
	output["data"] = mData

	return output, err
}
