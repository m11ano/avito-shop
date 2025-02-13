// Package dbhelper contains helper functions for converting
// between a database model (with tags) and a domain model (without tags).
package dbhelper

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// ConverterFunc defines a function type for converting a value from one type to another.
type ConverterFunc func(src interface{}) (interface{}, error)

var convertersMu sync.RWMutex

// converters is a registry of converter functions.
// The key is formatted as "srcType->dstType".
var converters = make(map[string]ConverterFunc)

// converterKey creates a registry key for a converter function.
func converterKey(srcType, dstType reflect.Type) string {
	return srcType.String() + "->" + dstType.String()
}

// RegisterBidirectionalConverter registers a pair of conversion functions for converting
// between a database field type and a domain model field type.
// Parameters:
//   - dbType: the type as stored in the database;
//   - domainType: the type in the domain model;
//   - dbToDomain: function to convert from dbType to domainType;
//   - domainToDB: function to convert from domainType to dbType.
func RegisterBidirectionalConverter(dbType, domainType reflect.Type, dbToDomain ConverterFunc, domainToDB ConverterFunc) {
	convertersMu.Lock()
	defer convertersMu.Unlock()
	converters[converterKey(dbType, domainType)] = dbToDomain
	converters[converterKey(domainType, dbType)] = domainToDB
}

// getConverter retrieves the registered converter function for converting from srcType to dstType.
func getConverter(srcType, dstType reflect.Type) (ConverterFunc, bool) {
	convertersMu.RLock()
	defer convertersMu.RUnlock()
	conv, ok := converters[converterKey(srcType, dstType)]
	return conv, ok
}

// safeConvert calls a converter function and catches any panic that occurs during its execution.
// If a panic is caught, it returns an error with the panic information.
func safeConvert(conv ConverterFunc, src interface{}) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in converter: %v", r)
		}
	}()
	result, err = conv(src)
	return
}

// snakeToCamel converts a snake_case string to CamelCase.
// For example, "birth_date" becomes "BirthDate".
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// lookupField attempts to find a field in typ corresponding to a given column name.
// It first tries an exact match, then converts the column name from snake_case to CamelCase.
// If the result ends with "Id", it also tries replacing it with "ID".
func lookupField(typ reflect.Type, colName string) (reflect.StructField, bool) {
	// Try an exact match using the column name.
	if field, ok := typ.FieldByName(colName); ok {
		return field, true
	}
	// Try converting from snake_case to CamelCase.
	camel := snakeToCamel(colName)
	if field, ok := typ.FieldByName(camel); ok {
		return field, true
	}
	// If the camel name ends with "Id", try replacing it with "ID".
	if strings.HasSuffix(camel, "Id") {
		altCamel := camel[:len(camel)-2] + "ID"
		if field, ok := typ.FieldByName(altCamel); ok {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

// ConvertDBToDomain converts a database record (dbRecord) into a domain model (domainModel).
// The database model provides the mapping between columns and fields via struct tags (e.g., db:"birth_date").
// The domain model does not contain tags, so the function first attempts to match by tag name and then
// converts snake_case names to CamelCase (including handling "Id" vs "ID").
// domainModel must be a pointer to a struct.
func ConvertDBToDomain(dbRecord, domainModel interface{}) error {
	// Get the value of the database record (struct or pointer to struct).
	dbVal := reflect.ValueOf(dbRecord)
	if dbVal.Kind() == reflect.Ptr {
		dbVal = dbVal.Elem()
	}
	if dbVal.Kind() != reflect.Struct {
		return fmt.Errorf("dbRecord must be a struct or pointer to a struct")
	}

	// domainModel must be a pointer to a struct.
	dVal := reflect.ValueOf(domainModel)
	if dVal.Kind() != reflect.Ptr || dVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("domainModel must be a pointer to a struct")
	}
	dVal = dVal.Elem()
	dType := dVal.Type()

	dbType := dbVal.Type()
	for i := 0; i < dbType.NumField(); i++ {
		dbField := dbType.Field(i)
		dbFieldValue := dbVal.Field(i)
		if !dbFieldValue.IsValid() {
			continue
		}
		// Extract the column name from the tag; if absent, use the field name.
		colName := dbField.Tag.Get("db")
		if colName == "" {
			colName = dbField.Name
		}

		// Look up the corresponding field in the domain model.
		dField, found := lookupField(dType, colName)
		if !found {
			continue // No matching field found; skip.
		}
		dFieldVal := dVal.FieldByName(dField.Name)
		if !dFieldVal.CanSet() {
			continue
		}

		// If types are identical or assignable, set the value directly.
		if dbFieldValue.Type().AssignableTo(dFieldVal.Type()) {
			dFieldVal.Set(dbFieldValue)
		} else if dbFieldValue.Type().ConvertibleTo(dFieldVal.Type()) {
			dFieldVal.Set(dbFieldValue.Convert(dFieldVal.Type()))
		} else {
			// If types differ, try using a registered converter.
			conv, found := getConverter(dbFieldValue.Type(), dFieldVal.Type())
			if !found {
				return fmt.Errorf("no converter registered for field %s: %s -> %s", dbField.Name, dbFieldValue.Type(), dFieldVal.Type())
			}
			converted, err := safeConvert(conv, dbFieldValue.Interface())
			if err != nil {
				return fmt.Errorf("converter error for field %s: %v", dbField.Name, err)
			}
			convVal := reflect.ValueOf(converted)
			if !convVal.Type().AssignableTo(dFieldVal.Type()) {
				return fmt.Errorf("converted value for field %s is not assignable to type %s", dbField.Name, dFieldVal.Type())
			}
			dFieldVal.Set(convVal)
		}
	}
	return nil
}

// StructToDBMap converts any structure (e.g., a domain model) into a map[string]interface{}
// for storing in the database. The schema is provided by a database model (dbSchema).
// For the map's keys, the function uses the value from the "db" tag (if present) from dbSchema.
// The corresponding field value is fetched from src using the original field name from dbSchema.
func StructToDBMap(src, dbSchema interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Get the database schema (struct or pointer to struct).
	schemaVal := reflect.ValueOf(dbSchema)
	if schemaVal.Kind() == reflect.Ptr {
		schemaVal = schemaVal.Elem()
	}
	if schemaVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("dbSchema must be a struct or pointer to a struct")
	}
	schemaType := schemaVal.Type()

	// Get the source (domain model) (struct or pointer to struct).
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	if srcVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("src must be a struct or pointer to a struct")
	}
	srcType := srcVal.Type()

	// Iterate over each field in the DB schema.
	for i := 0; i < schemaType.NumField(); i++ {
		schemaField := schemaType.Field(i)
		// Use the value from the "db" tag as the key; if not present, use the field's name.
		key := schemaField.Tag.Get("db")
		if key == "" {
			key = schemaField.Name
		}

		// Find the corresponding field in the source by using the original field name.
		srcField, found := srcType.FieldByName(schemaField.Name)
		if !found {
			continue // Field not found in source; skip.
		}
		srcFieldVal := srcVal.FieldByName(srcField.Name)
		expectedType := schemaField.Type

		// Direct assignment if types match.
		if srcFieldVal.Type().AssignableTo(expectedType) {
			result[key] = srcFieldVal.Interface()
		} else if srcFieldVal.Type().ConvertibleTo(expectedType) {
			// Use conversion if possible.
			result[key] = srcFieldVal.Convert(expectedType).Interface()
		} else {
			// Use a registered converter if available.
			conv, found := getConverter(srcFieldVal.Type(), expectedType)
			if !found {
				continue // No converter found; skip this field.
			}
			converted, err := safeConvert(conv, srcFieldVal.Interface())
			if err != nil {
				continue // Conversion error; skip.
			}
			convVal := reflect.ValueOf(converted)
			if !convVal.Type().AssignableTo(expectedType) {
				continue
			}
			result[key] = converted
		}
	}

	return result, nil
}

// ExtractDBFields extracts a slice of database field names from the given dbModel.
// dbModel must be a struct or a pointer to a struct. For each field, the function checks the "db" tag.
// If a "db" tag is present, its value is used; otherwise, the field's name is used.
// If dbModel is not a struct or pointer to a struct, an empty slice is returned.
func ExtractDBFields(dbModel interface{}) []string {
	val := reflect.ValueOf(dbModel)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return []string{}
	}
	typ := val.Type()
	var fields []string
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = field.Name
		}
		fields = append(fields, dbTag)
	}
	return fields
}
