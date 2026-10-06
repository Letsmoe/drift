package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var errorType = reflect.TypeFor[error]()

// dispatcher calls the exported methods of one value by name. Arguments arrive as
// a JSON array in parameter order; a method returns nothing, a value, an error, or
// a value and an error.
type dispatcher struct {
	target  reflect.Value
	methods map[string]reflect.Method
}

func newDispatcher(target any) *dispatcher {
	value := reflect.ValueOf(target)
	methods := make(map[string]reflect.Method)
	for index := range value.Type().NumMethod() {
		method := value.Type().Method(index)
		methods[method.Name] = method
	}
	return &dispatcher{target: value, methods: methods}
}

// call runs method with params and returns its value (nil when it has none).
func (dispatcher *dispatcher) call(name string, params []json.RawMessage) (result any, err error) {
	method, found := dispatcher.methods[name]
	if !found {
		return nil, fmt.Errorf("unknown method %q", name)
	}
	arguments, err := decodeArguments(method, params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}

	defer func() {
		recovered := recover()
		if recovered != nil {
			result = nil
			err = fmt.Errorf("%s panicked: %v", name, recovered)
		}
	}()
	outputs := dispatcher.target.Method(method.Index).Call(arguments)
	return splitOutputs(outputs)
}

func decodeArguments(method reflect.Method, params []json.RawMessage) ([]reflect.Value, error) {
	// The method type includes the receiver as its first input.
	parameterCount := method.Type.NumIn() - 1
	if len(params) != parameterCount {
		return nil, fmt.Errorf("takes %d arguments, got %d", parameterCount, len(params))
	}
	arguments := make([]reflect.Value, parameterCount)
	for index, param := range params {
		argument := reflect.New(method.Type.In(index + 1))
		err := json.Unmarshal(param, argument.Interface())
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", index+1, err)
		}
		arguments[index] = argument.Elem()
	}
	return arguments, nil
}

func splitOutputs(outputs []reflect.Value) (any, error) {
	if len(outputs) == 0 {
		return nil, nil
	}
	last := outputs[len(outputs)-1]
	if last.Type() != errorType {
		return outputs[0].Interface(), nil
	}
	if !last.IsNil() {
		return nil, last.Interface().(error) //nolint:forcetypeassert // checked against errorType above
	}
	if len(outputs) == 1 {
		return nil, nil
	}
	return outputs[0].Interface(), nil
}
