package metadata

import "strconv"

type Map map[string]interface{}

func (m Map) AsInt(key string, defaultValue int) int {
	if v, ok := m[key]; ok {
		switch vv := v.(type) {
		case string:
			n, err := strconv.Atoi(vv)
			if err != nil {
				return 0
			}
			return n
		case float64:
			return int(vv)
		case int:
			return vv
		case int64:
			return int(vv)
		}
	}
	return defaultValue
}

func (m Map) AsString(key string, defaultValue string) string {
	switch v := m[key].(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', 6, 64)
	}
	return defaultValue
}

func (m Map) AsMap(key string) Map {
	if m == nil {
		return Map{}
	}
	if v, ok := m[key]; ok {
		switch vv := v.(type) {
		case map[string]interface{}:
			return vv
		case Map:
			return vv
		case *Map:
			if vv != nil {
				return *vv
			}
		}
	}
	return Map{}
}

func (m Map) AsFloat64(key string, defaultValue float64) float64 {
	if v, ok := m[key]; ok {
		switch vv := v.(type) {
		case float64:
			return vv
		case int:
			return float64(vv)
		case int64:
			return float64(vv)
		case string:
			vvv, err := strconv.ParseFloat(vv, 64)
			if err == nil {
				return vvv
			}
			return 0.0
		}
	}
	return defaultValue
}
