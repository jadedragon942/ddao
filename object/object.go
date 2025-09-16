package object

import (
	"fmt"
	"log"
)

type Object struct {
	TableName string
	ID        string // Unique identifier for the object
	Fields    map[string]any
}

func New() *Object {
	return &Object{
		TableName: "",
		Fields:    make(map[string]any),
	}
}

func (o *Object) GetField(fieldName string) (any, bool) {
	value, exists := o.Fields[fieldName]
	return value, exists
}

func (o *Object) SetField(fieldName string, value any) {
	o.Fields[fieldName] = value
}

func (o *Object) GetTableName() string {
	return o.TableName
}

func (o *Object) SetTableName(tableName string) {
	o.TableName = tableName
}

func (o *Object) GetFields() map[string]any {
	return o.Fields
}

func (o *Object) SetFields(fields map[string]any) {
	o.Fields = fields
}

func (o *Object) GetString(fieldName string) (string, bool) {
	value, exists := o.Fields[fieldName]
	if !exists {
		return "", false
	}
	switch value.(type) {
	case *string:
		// If the value is a pointer to a string, dereference it
		if strValue, ok := value.(*string); ok {
			return *strValue, true
		}
	case string:
		// If the value is already a string, return it directly
		return value.(string), true
	default:
		// If the value is of an unexpected type, return an empty string and false
		log.Printf("Unexpected type for field %s: %T", fieldName, value)
		return "", false
	}
	return "", false
}

func (o *Object) GetInt64(fieldName string) (int64, bool) {
	value, exists := o.Fields[fieldName]
	if !exists {
		return 0, false
	}
	switch value.(type) {
	case int64:
		// If the value is already an int64, return it directly
		return value.(int64), true
	case int:
		// If the value is an int, convert it to int64
		return int64(value.(int)), true
	case string:
		// If the value is a string, try to convert it to int64
		var intValue int64
		_, err := fmt.Sscanf(value.(string), "%d", &intValue)
		if err != nil {
			log.Printf("Error converting string to int64 for field %s: %v", fieldName, err)
			return 0, false
		}
		return intValue, true
	case []byte:
		// If the value is a byte slice, try to convert it to int64
		var intValue int64
		_, err := fmt.Sscanf(string(value.([]byte)), "%d", &intValue)
		if err != nil {
			log.Printf("Error converting byte slice to int64 for field %s: %v", fieldName, err)
			return 0, false
		}
		return intValue, true
	case float64:
		// If the value is a float64, convert it to int64
		return int64(value.(float64)), true
	case float32:
		// If the value is a float32, convert it to int64
		return int64(value.(float32)), true
	case uint64:
		// If the value is a uint64, convert it to int64
		return int64(value.(uint64)), true
	case uint32:
		// If the value is a uint32, convert it to int64
		return int64(value.(uint32)), true
	default:
		// If the value is of an unexpected type, return an empty string and false
		log.Printf("Unexpected type for field %s: %T", fieldName, value)
		return 0, false
	}
}

func (o *Object) GetInt(fieldName string) (int, bool) {
	value, exists := o.Fields[fieldName]
	if !exists {
		return 0, false
	}
	intValue, ok := value.(int)
	return intValue, ok
}

func (o *Object) GetUint64(fieldName string) (uint64, bool) {
	value, exists := o.Fields[fieldName]
	if !exists {
		return 0, false
	}
	uint64Value, ok := value.(uint64)
	return uint64Value, ok
}

func (o *Object) GetFloat64(fieldName string) (float64, bool) {
	value, exists := o.Fields[fieldName]
	if !exists {
		return 0.0, false
	}
	floatValue, ok := value.(float64)
	return floatValue, ok
}

func (o *Object) GetBool(fieldName string) (bool, bool) {
	value, exists := o.Fields[fieldName]
	if !exists {
		return false, false
	}
	boolValue, ok := value.(bool)
	return boolValue, ok
}
