package testutils

import (
	"reflect"
)

const maxDepth = 32

// MergeMap recursively merges the src and dst maps. Key conflicts are resolved by
// preferring src, or recursively descending, if both src and dst are maps.
func MergeMap(dst, src map[string]any) map[string]any {
	return merge(dst, src, 0)
}

func merge(dst, src map[string]any, depth int) map[string]any {
	if depth > maxDepth {
		panic("too deep!")
	}
	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			srcMap, srcMapOk := mapify(srcVal)
			dstMap, dstMapOk := mapify(dstVal)
			if srcMapOk && dstMapOk {
				srcVal = merge(dstMap, srcMap, depth+1)
			}
		}
		dst[key] = srcVal
	}
	return dst
}

func mapify(i any) (map[string]any, bool) {
	switch v := i.(type) {
	case map[string]any:
		return v, true
	case Values:
		return v, true
	}

	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := make(map[string]any, value.Len())
		for _, k := range value.MapKeys() {
			m[k.String()] = value.MapIndex(k).Interface()
		}
		return m, true
	}
	return map[string]any{}, false
}
