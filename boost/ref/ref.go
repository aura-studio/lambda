package ref

import "reflect"

func TypeName(i interface{}) string {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		return reflect.TypeOf(i).Elem().Name()
	} else if v.Kind() == reflect.Struct {
		return reflect.TypeOf(i).Name()
	}
	return ""
}

func CallMethod(i interface{}, method string) []interface{} {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct || v.Kind() == reflect.Struct {
		var results []interface{}
		values := reflect.ValueOf(i).MethodByName(method).Call([]reflect.Value{})
		for _, value := range values {
			results = append(results, value.Interface())
		}
		return results
	}
	return nil
}
