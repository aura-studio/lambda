package tree

import (
	"reflect"
	"strconv"

	"github.com/mohae/deepcopy"
)

type Tree struct {
	data map[string]interface{}
}

func NewTree() *Tree {
	return &Tree{
		data: make(map[string]interface{}),
	}
}

func (t *Tree) SetData(data map[string]interface{}) *Tree {
	t.data = data
	return t
}

func (t *Tree) Data() interface{} {
	return t.data
}

func (t *Tree) Get(path []string) interface{} {
	return t.get(t.data, path)
}

func (t *Tree) get(source interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	switch source := source.(type) {
	case map[string]interface{}:
		next, ok := source[path[0]]
		if !ok {
			return nil
		}
		return t.get(next, path[1:])
	case []interface{}:
		index, err := strconv.Atoi(path[0])
		if err != nil || len(source) <= index {
			return nil
		}
		next := source[index]
		return t.get(next, path[1:])
	default:
		return nil
	}
}

func (t *Tree) Set(path []string, v interface{}) {
	t.data = t.set(t.data, path, v).(map[string]interface{})
}

func (t *Tree) set(source interface{}, path []string, v interface{}) interface{} {
	if len(path) == 0 {
		return source
	}

	if len(path) == 1 {
		switch source := source.(type) {
		case map[string]interface{}:
			source[path[0]] = v
			return source
		case []interface{}:
			index, err := strconv.Atoi(path[0])
			if err != nil {
				return nil
			}
			for len(source) <= index {
				source = append(source, nil)
			}
			source[index] = v
			return source
		default:
			return source
		}
	}

	switch source := source.(type) {
	case map[string]interface{}:
		next, ok := source[path[0]]
		if !ok {
			_, err := strconv.Atoi(path[1])
			if err != nil {
				next = make(map[string]interface{})
			} else {
				next = make([]interface{}, 0)
			}
		}
		source[path[0]] = t.set(next, path[1:], v)
		return source
	case []interface{}:
		index, err := strconv.Atoi(path[0])
		if err != nil {
			return source
		}
		for len(source) <= index {
			source = append(source, nil)
		}
		next := source[index]
		if next == nil {
			_, err := strconv.Atoi(path[1])
			if err != nil {
				next = make(map[string]interface{})
			} else {
				next = make([]interface{}, 0)
			}
		}
		source[index] = t.set(next, path[1:], v)
		return source
	default:
		return source
	}
}

func (t *Tree) Remove(path []string) {
	t.remove(t.data, path)
}

func (t *Tree) remove(source interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	if len(path) == 1 {
		switch source := source.(type) {
		case map[string]interface{}:
			source[path[0]] = nil
			for k, v := range source {
				if v == nil {
					delete(source, k)
				}
			}
			if len(source) == 0 {
				return nil
			}
			return source
		case []interface{}:
			index, err := strconv.Atoi(path[0])
			if err != nil || len(source) <= index {
				return source
			}
			source[index] = nil
			for i := len(source) - 1; i >= 0; i-- {
				if source[i] == nil {
					source = source[:len(source)-1]
				} else {
					break
				}
			}
			if len(source) == 0 {
				return nil
			}
			return source
		default:
			return source
		}
	}

	switch source := source.(type) {
	case map[string]interface{}:
		next, ok := source[path[0]]
		if !ok {
			return source
		}
		source[path[0]] = t.remove(next, path[1:])
		for k, v := range source {
			if v == nil {
				delete(source, k)
			}
		}
		if len(source) == 0 {
			return nil
		}
		return source
	case []interface{}:
		index, err := strconv.Atoi(path[0])
		if err != nil || len(source) <= index {
			return source
		}
		next := source[index]
		source[index] = t.remove(next, path[1:])
		for i := len(source) - 1; i >= 0; i-- {
			if source[i] == nil {
				source = source[:len(source)-1]
			} else {
				break
			}
		}
		if len(source) == 0 {
			return nil
		}
		return source
	default:
		return source
	}
}

func (t *Tree) Merge(smt *Tree) {
	t.data = t.merge(smt.data, t.data).(map[string]interface{})
}

func (t *Tree) merge(source interface{}, target interface{}) interface{} {
	sourceType := reflect.TypeOf(source)
	targetType := reflect.TypeOf(target)
	if sourceType != targetType {
		return source
	}
	switch source := source.(type) {
	case map[string]interface{}:
		target := target.(map[string]interface{})
		for sk, sv := range source {
			dv, ok := target[sk]
			if !ok || dv == nil {
				target[sk] = sv
			} else {
				target[sk] = t.merge(sv, target[sk])
			}
		}
		return target
	case []interface{}:
		target := target.([]interface{})
		for index := 0; index < len(source); index++ {
			if index >= len(target) {
				target = append(target, source[index])
			} else {
				target[index] = t.merge(source[index], target[index])
			}
		}
		return target
	default:
		return source
	}
}

func (t *Tree) Dulplicate() *Tree {
	mapTree := NewTree()
	mapTree.SetData(deepcopy.Copy(t.data).(map[string]interface{}))
	return mapTree
}

func (t *Tree) MarshalHash() [][]interface{} {
	return t.marshalHash(make([][]interface{}, 0), t.data, []string{})
}

func (t *Tree) marshalHash(pairs [][]interface{}, source interface{}, prefix []string) [][]interface{} {
	switch source := source.(type) {
	case map[string]interface{}:
		for k, v := range source {
			pairs = t.marshalHash(pairs, v, append(prefix, k))
		}
	case []interface{}:
		for i := 0; i < len(source); i++ {
			pairs = t.marshalHash(pairs, source[i], append(prefix, strconv.Itoa(i)))
		}
	default:
		pairs = append(pairs, []interface{}{prefix, source})
	}
	return pairs
}

func (t *Tree) UnmarshalHash(pairs [][]interface{}) {
	for _, pair := range pairs {
		var ss []string
		for _, s := range pair[0].([]interface{}) {
			ss = append(ss, s.(string))
		}

		t.Set(ss, pair[1])
	}
}
