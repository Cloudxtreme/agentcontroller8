package jswatcher

import (
	"fmt"
)

/*
ATTRIBUTES = ('descr', 'organization', 'name', 'author',
              'license', 'version', 'category', 'async',
              'queue', 'roles', 'enable', 'period', 'timeout')
*/

type jsAttributes map[string]interface{}

func (attr jsAttributes) getInt(key string, default_ int) (int, error) {
	valueI, ok := attr[key]
	if !ok {
		return default_, nil
	}

	switch t := valueI.(type) {
	case int:
		return t, nil
	case float64:
		return int(t), nil
	default:
		return default_, fmt.Errorf("Invalid type for '%s'", valueI)
	}
}

func (attr jsAttributes) getBool(key string, default_ bool) (bool, error) {
	valueI, ok := attr[key]
	if !ok {
		return default_, nil
	}

	value, ok := valueI.(bool)
	if !ok {
		return default_, fmt.Errorf("Invalid type for '%s'", value)
	}

	return value, nil
}

func (attr jsAttributes) getString(key string, default_ string) (string, error) {
	valueI, ok := attr[key]
	if !ok {
		return default_, nil
	}

	value, ok := valueI.(string)
	if !ok {
		return default_, fmt.Errorf("Invalid type for '%s'", value)
	}

	return value, nil
}

func (attr jsAttributes) getStringList(key string) ([]string, error) {
	result := make([]string, 0)

	valueI, ok := attr[key]
	if !ok {
		return result, nil
	}

	value, ok := valueI.([]interface{})
	if !ok {
		return result, fmt.Errorf("Invalid type for '%s'", value)
	}

	for _, v := range value {
		vstr, ok := v.(string)
		if !ok {
			return result, fmt.Errorf("Invalud type for '%s'", v)
		}

		result = append(result, vstr)
	}

	return result, nil
}

func (attr jsAttributes) Period() int {
	value, _ := attr.getInt("period", 0)
	return value
}

func (attr jsAttributes) Timeout() int {
	value, _ := attr.getInt("timeout", 0)
	return value
}

func (attr jsAttributes) Enable() bool {
	value, _ := attr.getBool("enable", true)
	return value
}

func (attr jsAttributes) Queue() string {
	value, _ := attr.getString("queue", "")
	return value
}

func (attr jsAttributes) Category() string {
	value, _ := attr.getString("category", "")
	return value
}

func (attr jsAttributes) Roles() []string {
	roles, _ := attr.getStringList("roles")
	return roles
}

func (attr jsAttributes) String() string {
	return fmt.Sprintf("Enable: %v, Period: %v, Timeout: %v, Queue: %v, Category: %v, Roles: %v",
		attr.Enable(), attr.Period(), attr.Timeout(), attr.Queue(), attr.Category(), attr.Roles())
}
