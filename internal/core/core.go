package core

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/metadata"
	logger "github.com/project-flogo/core/support/log"
)

type microgatewayHost struct {
	id         string
	name       string
	scope      data.Scope
	iometadata *metadata.IOMetadata
	err        error
	halt       bool
}

func (m *microgatewayHost) ID() string {
	return m.id
}

func (m *microgatewayHost) Name() string {
	return m.name
}

func (m *microgatewayHost) IOMetadata() *metadata.IOMetadata {
	return m.iometadata
}

func (m *microgatewayHost) Reply(replyData map[string]interface{}, err error) {
	for key, value := range replyData {
		m.scope.SetValue(key, value)
	}
	m.err = err
}

func (m *microgatewayHost) Return(returnData map[string]interface{}, err error) {
	m.Reply(returnData, err)
	m.halt = true
}

func (m *microgatewayHost) Scope() data.Scope {
	return m.scope
}

// Execute executes the microgateway
func Execute(id string, payload interface{}, definition *Microgateway, iometadata *metadata.IOMetadata, log logger.Logger) (code int, output interface{}, err error) {

	// Contains all elements of request: right now just payload, environment flags and service instances.
	envFlags := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		envFlags[pair[0]] = pair[1]
	}
	executionContext := map[string]interface{}{
		"payload": payload,
		"async":   definition.Async,
		"env":     envFlags,
		"conf":    definition.Configuration,
	}
	scope := data.NewSimpleScope(executionContext, nil)
	host := microgatewayHost{
		id:         id,
		name:       definition.Name,
		scope:      scope,
		iometadata: iometadata,
	}

	// Execute the identified route if it exists and handle the async option.
	if definition != nil {
		if definition.Async {
			log.Info("executing route asynchronously")
			go func() {
				done, err := executeSteps(definition, &host, log)
				if err != nil {
					if done {
						log.Info("error executing route: ", err)
					} else {
						log.Error("error executing route: ", err)
					}
				}
			}()
		} else {
			var done bool
			done, err = executeSteps(definition, &host, log)
			if err != nil {
				if done {
					log.Info("error executing route: ", err)
				} else {
					log.Error("error executing route: ", err)
				}
			}
		}
	} else {
		log.Info("no route to execute, continuing to reply handler")
	}

	if definition != nil {
		for _, response := range definition.Responses {
			var truthiness bool
			truthiness, err = evaluateTruthiness(response.Condition, scope, log)
			if err != nil {
				continue
			}
			if truthiness {
				output, oErr := TranslateMappings(scope, map[string]*Expr{"code": response.Output.Code}, log)
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
				if response.Output.Datum != nil {
					data, oErr = TranslateMappings(scope, response.Output.Datum, log)
					if oErr != nil {
						return -1, nil, oErr
					}
				} else {
					interimData, dErr := TranslateMappings(scope, map[string]*Expr{"data": response.Output.Data}, log)
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

func executeSteps(definition *Microgateway, host *microgatewayHost, log logger.Logger) (done bool, err error) {
	for _, step := range definition.Steps {
		var truthiness bool
		truthiness, err = evaluateTruthiness(step.Condition, host.Scope(), log)
		if err != nil {
			continue
		}
		ctxt := newServiceContext(step.Service, host, log)
		ctxt.UpdateScope(nil)
		if truthiness {
			done, err = invokeService(step.Service, step.HaltCondition, host, ctxt, step.Input, log)
			if err != nil {
				return done, err
			}
			if host.halt {
				return true, nil
			}
		}
	}
	return true, nil
}

func evaluateTruthiness(expr *Expr, scope data.Scope, log logger.Logger) (truthy bool, err error) {
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
	host    activity.Host
	logger  logger.Logger
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}

func newServiceContext(def *Service, host activity.Host, log logger.Logger) *serviceContext {
	inputs := make(map[string]interface{}, len(def.Settings))
	for k, v := range def.Settings {
		inputs[k] = v
	}
	return &serviceContext{
		name:    def.Name,
		host:    host,
		logger:  logger.ChildLogger(log, def.Name),
		Inputs:  inputs,
		Outputs: make(map[string]interface{}),
	}
}

func (s *serviceContext) Merge(inputs map[string]interface{}) {
	for k, v := range inputs {
		s.Inputs[k] = v
	}
}

func (s *serviceContext) UpdateScope(err error) {
	activityData := map[string]interface{}{
		"inputs":  s.Inputs,
		"outputs": s.Outputs,
		"error":   err,
	}
	s.host.Scope().SetValue(s.name, activityData)
}

func (s *serviceContext) ActivityHost() activity.Host {
	return s.host
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

func (s *serviceContext) Logger() logger.Logger {
	return s.logger
}

func invokeService(serviceDef *Service, haltCondition *Expr, host *microgatewayHost, ctxt *serviceContext, input map[string]*Expr, log logger.Logger) (done bool, err error) {
	log.Info("invoking service: ", serviceDef.Name)

	scope := host.Scope()
	values, err := TranslateMappings(scope, input, log)
	if err != nil {
		return false, err
	}
	ctxt.Merge(values)

	ctxt.UpdateScope(nil)
	done, err = serviceDef.Activity.Eval(ctxt)

	if err == nil {
		err = host.err
	}
	ctxt.UpdateScope(err)
	if haltCondition != nil {
		truthiness, err := evaluateTruthiness(haltCondition, scope, log)
		if err != nil {
			return true, nil
		}
		if truthiness {
			return true, fmt.Errorf("execution halted with expression: %s", haltCondition)
		}
		return false, nil
	}

	return done, err
}

// TranslateMappings translates dot notation mappings
func TranslateMappings(scope data.Scope, mappings map[string]*Expr, log logger.Logger) (values map[string]interface{}, err error) {
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
