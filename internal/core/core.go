package core

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/logger"
	"github.com/project-flogo/microgateway/internal/types"
)

var log = logger.GetLogger("microgateway")

func Execute(payload interface{}, configuration map[string]interface{}, routes []types.Route,
	serviceCache map[string]*types.Service) (code int, output interface{}, err error) {
	// Route to be executed once it is identified by the conditional evaluation.
	var routeToExecute *types.Route

	// Contains all elements of request: right now just payload, environment flags and service instances.
	envFlags := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		envFlags[pair[0]] = pair[1]
	}
	executionContext := map[string]interface{}{
		"payload": payload,
		"async":   false,
		"env":     envFlags,
		"conf":    configuration,
	}
	scope := data.NewSimpleScope(executionContext, nil)

	// Evaluate route conditions to select which one to execute.
	for _, route := range routes {
		var truthiness bool
		truthiness, err = evaluateTruthiness(route.Expression, scope)
		if err != nil {
			continue
		}
		if truthiness {
			log.Info("route identified via conditional evaluation to true: ", route.Condition)
			routeToExecute = &route
			break
		}
	}

	// Execute the identified route if it exists and handle the async option.
	if routeToExecute != nil {
		if routeToExecute.Async {
			log.Info("executing route asynchronously")
			scope.SetValue("async", true)
			go executeRoute(routeToExecute, serviceCache, scope)
		} else {
			err = executeRoute(routeToExecute, serviceCache, scope)
		}
		if err != nil {
			log.Error("error executing route: ", err)
		}
	} else {
		log.Info("no route to execute, continuing to reply handler")
	}

	if routeToExecute != nil {
		for _, response := range routeToExecute.Responses {
			var truthiness bool
			truthiness, err = evaluateTruthiness(response.Expression, scope)
			if err != nil {
				continue
			}
			if truthiness {
				output, oErr := translateMappings(scope, map[string]*types.Expr{"code": response.Output.CodeExpression})
				if oErr != nil {
					return -1, nil, oErr
				}
				var code int
				codeElement, ok := output["code"]
				if ok {
					switch cv := codeElement.(type) {
					case float64:
						code = int(cv)
					case int:
						code = cv
					case string:
						code, err = strconv.Atoi(cv)
						if err != nil {
							log.Info("unable to format extracted code string from response output", cv)
						}
					}
				}
				if ok && code != 0 {
					log.Info("Code identified in response output: ", code)
				} else {
					log.Info("Code contents is not found or not an integer, default response code is 200")
					code = 200
				}
				// Translate data mappings
				var data interface{}
				if response.Output.DataExpressions != nil {
					data, oErr = translateMappings(scope, response.Output.DataExpressions)
					if oErr != nil {
						return -1, nil, oErr
					}
				} else {
					interimData, dErr := translateMappings(scope, map[string]*types.Expr{"data": response.Output.DataExpression})
					if dErr != nil {
						return -1, nil, dErr
					}
					data, ok = interimData["data"]
					if !ok {
						return -1, nil, errors.New("cannot extract data from response output")
					}
				}
				return code, data, err
			}
		}
	}
	return 404, nil, err
}

func executeRoute(route *types.Route, serviceCache map[string]*types.Service, scope data.Scope) (err error) {
	for _, step := range route.Steps {
		var truthiness bool
		truthiness, err = evaluateTruthiness(step.Expression, scope)
		if err != nil {
			return err
		}
		if truthiness {
			err = invokeService(serviceCache[step.Service], scope, step.InputExpression)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func evaluateTruthiness(expr *types.Expr, scope data.Scope) (truthy bool, err error) {
	if expr == nil {
		log.Info("condition was empty and thus evaluates to true")
		return true, nil
	}
	val, err := expr.Eval(scope)
	if err != nil {
		log.Infof("condition evaluation causes error so is false: %s", expr)
		return false, err
	}
	if val == nil {
		log.Infof("condition evaluation results in nil value so is false: %s", expr)
		return false, errors.New("expression has nil value")
	}
	truthy, ok := val.(bool)
	if !ok {
		log.Infof("condition evaluation results in non-bool value so is false: %s", expr)
		return false, errors.New("expression has a non-bool value")
	}

	log.Infof("condition evaluated to %t: %s", truthy, expr)
	return truthy, err
}

type serviceContext struct {
	name    string
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}

func newServiceContext(def *types.Service) *serviceContext {
	inputs := make(map[string]interface{}, len(def.Settings))
	for k, v := range def.Settings {
		inputs[k] = v
	}
	return &serviceContext{
		name:    def.Name,
		Inputs:  inputs,
		Outputs: make(map[string]interface{}),
	}
}

func (s *serviceContext) Merge(inputs map[string]interface{}) {
	for k, v := range inputs {
		s.Inputs[k] = v
	}
}

func (s *serviceContext) Context() map[string]interface{} {
	return map[string]interface{}{
		"inputs":  s.Inputs,
		"outputs": s.Outputs,
	}
}

func (s *serviceContext) ActivityHost() activity.Host {
	return s
}

func (s *serviceContext) Name() string {
	return s.name
}

func (s *serviceContext) GetInput(name string) interface{} {
	return s.Inputs[name]
}

func (s *serviceContext) SetOutput(name string, value interface{}) error {
	s.Outputs[name] = value
	return nil
}

func (s *serviceContext) GetInputObject(input data.StructValue) error {
	return input.FromMap(s.Inputs)
}

func (s *serviceContext) SetOutputObject(output data.StructValue) error {
	s.Outputs = output.ToMap()
	return nil
}

func (s *serviceContext) GetSharedTempData() map[string]interface{} {
	return nil
}

func (s *serviceContext) ID() string {
	return ""
}

func (s *serviceContext) IOMetadata() *metadata.IOMetadata {
	return nil
}

func (s *serviceContext) Reply(replyData map[string]interface{}, err error) {

}

func (s *serviceContext) Return(returnData map[string]interface{}, err error) {

}

func (s *serviceContext) Scope() data.Scope {
	return nil
}

func invokeService(serviceDef *types.Service, scope data.Scope, input map[string]*types.Expr) (err error) {
	log.Info("invoking service: ", serviceDef.Ref)
	// TODO: Translate service definition variables.
	ctxt := newServiceContext(serviceDef)
	defer func() {
		scope.SetValue(serviceDef.Name, ctxt.Context())
	}()
	scope.SetValue(serviceDef.Name, ctxt.Context())
	values, mErr := translateMappings(scope, input)
	if mErr != nil {
		return mErr
	}

	ctxt.Merge(values)
	_, err = serviceDef.Activity.Eval(ctxt)
	if err != nil {
		return err
	}
	return nil
}

func translateMappings(scope data.Scope, mappings map[string]*types.Expr) (values map[string]interface{}, err error) {
	values = make(map[string]interface{})
	if len(mappings) == 0 {
		return values, err
	}
	for fullKey, expr := range mappings {
		value, err := expr.Eval(scope)
		if err != nil {
			log.Infof("mapping evaluation causes error: %s", expr)
			return values, err
		}
		values[fullKey] = value
	}
	return expandMap(values), err
}

// Turn dot notation map into nested map structure.
func expandMap(m map[string]interface{}) map[string]interface{} {
	var tree = make(map[string]interface{})
	for key, value := range m {
		keys := strings.Split(key, ".")
		subTree := tree
		for _, treeKey := range keys[:len(keys)-1] {
			subTreeNew, ok := subTree[treeKey]
			if !ok {
				subTreeNew = make(map[string]interface{})
				subTree[treeKey] = subTreeNew
			}
			subTree = subTreeNew.(map[string]interface{})
		}
		subTree[keys[len(keys)-1]] = value
	}
	return tree
}
