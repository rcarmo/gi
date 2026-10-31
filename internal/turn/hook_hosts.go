package turn

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type hookHostGroup struct {
	hosts  sync.WaitGroup
	sealed bool
}

func hookOwner(req HookRequest) string {
	if req.TurnID != "" {
		return req.TurnID
	}
	for _, key := range []string{"turn_id", "active_turn_id"} {
		if id, ok := req.Payload[key].(string); ok {
			return id
		}
	}
	return ""
}
func (e *Engine) admitHookHost(req HookRequest) (*hookHostGroup, error) {
	e.hookMu.Lock()
	defer e.hookMu.Unlock()
	if e.hooksClosed {
		return nil, context.Canceled
	}
	var group *hookHostGroup
	if id := hookOwner(req); id != "" {
		if e.hookHosts == nil {
			e.hookHosts = map[string]*hookHostGroup{}
		}
		group = e.hookHosts[id]
		if group == nil {
			group = &hookHostGroup{}
			e.hookHosts[id] = group
		}
		if group.sealed {
			return nil, context.Canceled
		}
		group.hosts.Add(1)
	}
	e.hookRuns.Add(1)
	return group, nil
}
func (e *Engine) finishHookHost(group *hookHostGroup) {
	if group != nil {
		group.hosts.Done()
	}
	e.hookRuns.Done()
}
func (e *Engine) joinTurnHooks(turnID string, seal bool) {
	e.hookMu.Lock()
	group := e.hookHosts[turnID]
	if seal {
		if group == nil {
			if e.hookHosts == nil {
				e.hookHosts = map[string]*hookHostGroup{}
			}
			group = &hookHostGroup{}
			e.hookHosts[turnID] = group
		}
		group.sealed = true
	}
	e.hookMu.Unlock()
	if group != nil {
		group.hosts.Wait()
	}
}
func (e *Engine) forgetTurnHooks(turnID string) {
	e.hookMu.Lock()
	delete(e.hookHosts, turnID)
	e.hookMu.Unlock()
}
func (e *Engine) closeHookHosts() {
	e.hookMu.Lock()
	e.hooksClosed = true
	processes := e.hookProcesses
	e.hookProcesses = nil
	e.hookMu.Unlock()
	e.hookRuns.Wait()
	for process := range processes {
		process.close()
	}
}

// Hook envelopes carry data, not live capabilities. Preserve concrete scalar
// types (unlike JSON round trips), detach every mutable value, and fail closed
// on functions, channels, hidden state or excessively deep/cyclic graphs.
func detachHookEnvelope[T any](input T) (T, error) {
	var zero T
	cloned, err := detachHookValue(reflect.ValueOf(input), 0)
	if err != nil {
		return zero, err
	}
	return cloned.Interface().(T), nil
}
func detachHookValue(value reflect.Value, depth int) (reflect.Value, error) {
	if depth > 128 {
		return reflect.Value{}, fmt.Errorf("hook data is cyclic or exceeds 128 levels")
	}
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type()), nil
		}
		child, err := detachHookValue(value.Elem(), depth+1)
		if err != nil {
			return reflect.Value{}, err
		}
		out := reflect.New(value.Type()).Elem()
		out.Set(child)
		return out, nil
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type()), nil
		}
		child, err := detachHookValue(value.Elem(), depth+1)
		if err != nil {
			return reflect.Value{}, err
		}
		out := reflect.New(value.Type().Elem())
		out.Elem().Set(child)
		return out, nil
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type()), nil
		}
		out := reflect.MakeMapWithSize(value.Type(), value.Len())
		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key()
			switch key.Kind() {
			case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			default:
				return reflect.Value{}, fmt.Errorf("unsupported hook map key %s", key.Type())
			}
			child, err := detachHookValue(iter.Value(), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.SetMapIndex(key, child)
		}
		return out, nil
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type()), nil
		}
		out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			child, err := detachHookValue(value.Index(i), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.Index(i).Set(child)
		}
		return out, nil
	case reflect.Array:
		out := reflect.New(value.Type()).Elem()
		for i := 0; i < value.Len(); i++ {
			child, err := detachHookValue(value.Index(i), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.Index(i).Set(child)
		}
		return out, nil
	case reflect.Struct:
		out := reflect.New(value.Type()).Elem()
		for i := 0; i < value.NumField(); i++ {
			if value.Type().Field(i).PkgPath != "" {
				return reflect.Value{}, fmt.Errorf("hidden hook state in %s", value.Type())
			}
			child, err := detachHookValue(value.Field(i), depth+1)
			if err != nil {
				return reflect.Value{}, err
			}
			out.Field(i).Set(child)
		}
		return out, nil
	case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return value, nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported hook capability %s", value.Type())
	}
}
