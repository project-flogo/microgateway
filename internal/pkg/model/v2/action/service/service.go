package service

import (
	"encoding/json"
	"errors"

	//"github.com/TIBCOSoftware/mashling/internal/pkg/model/v2/action/service/grpc"
	//wsproxy "github.com/TIBCOSoftware/mashling/internal/pkg/model/v2/action/service/wsproxy"
	"github.com/dop251/goja"
	"github.com/project-flogo/microgateway/internal/pkg/model/v2/types"
	"github.com/project-flogo/microgateway/registry"
)

// Initialize sets up the service based off of the service definition.
func Initialize(serviceDef types.Service) (service registry.Service, err error) {
	factory := registry.Lookup(serviceDef.Type)
	if factory == nil {
		return nil, errors.New("unknown service type")
	}
	return factory.Make(serviceDef.Name, serviceDef.Settings)

	/*switch sType := serviceDef.Type; sType {
	case "http":
		return InitializeHTTP(serviceDef.Settings)
	case "js":
		return InitializeJS(serviceDef.Settings)
	case "flogoActivity":
		return InitializeFlogoActivity(serviceDef.Settings)
	case "flogoFlow":
		return InitializeFlogoFlow(serviceDef.Settings)
	case "sqld":
		return InitializeSQLD(serviceDef.Settings)
	case "grpc":
		return grpc.InitializeGRPC(serviceDef.Settings)
	case "circuitBreaker":
		return InitializeCircuitBreaker(serviceDef.Settings)
	case "anomaly":
		return InitializeAnomaly(serviceDef.Settings)
	case "jwt":
		return InitializeJWT(serviceDef.Settings)
	case "ws":
		return wsproxy.InitializeWSProxy(serviceDef.Name, serviceDef.Settings)
	case "ratelimiter":
		return InitializeRateLimiter(serviceDef.Name, serviceDef.Settings)
	default:
		return nil, errors.New("unknown service type")
	}*/
}

// VM represents a VM object.
type VM struct {
	vm *goja.Runtime
}

// NewVM initializes a new VM with defaults.
func NewVM(defaults map[string]interface{}) (vm *VM, err error) {
	vm = &VM{}
	vm.vm = goja.New()
	for k, v := range defaults {
		if v != nil {
			vm.vm.Set(k, v)
		}
	}
	return vm, err
}

// EvaluateToBool evaluates a string condition within the context of the VM.
func (vm *VM) EvaluateToBool(condition string) (truthy bool, err error) {
	if condition == "" {
		return true, nil
	}
	var res goja.Value
	res, err = vm.vm.RunString(condition)
	if err != nil {
		return false, err
	}
	truthy, ok := res.Export().(bool)
	if !ok {
		err = errors.New("condition does not evaluate to bool")
		return false, err
	}
	return truthy, err
}

// SetInVM sets the object name and value in the VM.
func (vm *VM) SetInVM(name string, object interface{}) (err error) {
	var valueJSON json.RawMessage
	var vmObject map[string]interface{}
	valueJSON, err = json.Marshal(object)
	if err != nil {
		return err
	}
	err = json.Unmarshal(valueJSON, &vmObject)
	if err != nil {
		return err
	}
	vm.vm.Set(name, vmObject)
	return err
}

// GetFromVM extracts the current object value from the VM.
func (vm *VM) GetFromVM(name string, object interface{}) (err error) {
	var valueJSON json.RawMessage
	var vmObject map[string]interface{}
	vm.vm.ExportTo(vm.vm.Get(name), &vmObject)

	valueJSON, err = json.Marshal(vmObject)
	if err != nil {
		return err
	}
	err = json.Unmarshal(valueJSON, object)
	if err != nil {
		return err
	}
	return err
}

// SetPrimitiveInVM sets primitive value in VM.
func (vm *VM) SetPrimitiveInVM(name string, primitive interface{}) {
	vm.vm.Set(name, primitive)
}
