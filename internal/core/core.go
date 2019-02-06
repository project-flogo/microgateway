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
				output, oErr := TranslateMappings(scope, []*Expr{response.Output.Code}, log)
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
					interimData, dErr := TranslateMappings(scope, []*Expr{response.Output.Data}, log)
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
	values  map[string]interface{}
}

func newServiceContext(def *Service, host activity.Host, log logger.Logger) *serviceContext {
	inputs := make(map[string]interface{}, len(def.Settings))
	for _, setting := range def.Settings {
		inputs[setting.Name] = setting.Value
	}
	ctxt := &serviceContext{
		name:    def.Name,
		host:    host,
		logger:  logger.ChildLogger(log, def.Name),
		Inputs:  inputs,
		Outputs: make(map[string]interface{}),
	}
	ctxt.values = map[string]interface{}{
		"inputs":  inputs,
		"outputs": ctxt.Outputs,
		"error":   nil,
	}
	host.Scope().SetValue(def.Name, ctxt.values)
	return ctxt
}

func (s *serviceContext) SetError(err error) {
	s.values["error"] = err
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
	s.values["outputs"] = s.Outputs
	return nil
}

func (s *serviceContext) GetSharedTempData() map[string]interface{} {
	return nil
}

func (s *serviceContext) Logger() logger.Logger {
	return s.logger
}

func invokeService(serviceDef *Service, haltCondition *Expr, host *microgatewayHost, ctxt *serviceContext, input []*Expr, log logger.Logger) (done bool, err error) {
	log.Info("invoking service: ", serviceDef.Name)

	scope := host.Scope()
	err = TranslateMappingsToTree(scope, input, ctxt.Inputs, log)
	if err != nil {
		return false, err
	}

	done, err = serviceDef.Activity.Eval(ctxt)
	if err == nil {
		err = host.err
	}
	ctxt.SetError(err)

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
func TranslateMappings(scope data.Scope, mappings []*Expr, log logger.Logger) (tree map[string]interface{}, err error) {
	length := len(mappings)
	tree = make(map[string]interface{}, length)
	if length == 0 {
		return tree, err
	}
	err = TranslateMappingsToTree(scope, mappings, tree, log)
	return tree, err
}

// TranslateMappingsToTree translates dot notation mappings
func TranslateMappingsToTree(scope data.Scope, mappings []*Expr, tree map[string]interface{}, log logger.Logger) (err error) {
	for _, expr := range mappings {
		value, err := expr.Eval(scope)
		if err != nil {
			log.Infof("mapping evaluation causes error: %s", expr)
			return err
		}

		keys, subTree := strings.Split(expr.Name, "."), tree
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
	return err
}
