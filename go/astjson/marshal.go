// Package astjson provides JSON serialization for the Go sqlparser AST.
// It uses reflection to walk the AST tree and produce JSON output.
package astjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unsafe"
)

// baseTypes lists type names that are base/marker implementations to skip.
var baseTypes = map[string]bool{
	"BaseNode":       true,
	"BaseStatement":  true,
	"ExpressionBase": true,
	"DataTypeBase":   true,
	"QueryBase":      true,
	"SBaseStatement": true,
	"BaseSetExpr":    true,
	"SelectItemBase": true,
}

// fields to skip in serialization - source location info.
var skipFields = map[string]bool{
	"SpanVal": true,
	"span":    true,
}

// statementTypes maps known Go statement type names to their JSON tag names.
var statementTypes = map[string]string{
	"SelectStatement": "Query",
	"QueryStatement":  "Query",
	"SCreateTable":    "CreateTable",
	"SQuery":          "Query",
	"SInsert":         "Insert",
	"SDelete":         "Delete",
	"SUpdate":         "Update",
	"SMerge":          "Merge",
	"SDrop":           "Drop",
	"SCreateView":     "CreateView",
	"SCreateIndex":    "CreateIndex",
	"SAlterTable":     "AlterTable",
	"SSet":            "Set",
	"SExplain":        "Explain",
	"SUse":            "Use",
	"SGrant":          "Grant",
	"SRevoke":         "Revoke",
	"SStartTransaction": "StartTransaction",
	"SCommit":         "Commit",
	"SRollback":       "Rollback",
	"SSavepoint":      "Savepoint",
	"Flush":           "Flush",
	"Kill":            "Kill",
	"CreateTable":     "CreateTable",
	"Insert":          "Insert",
	"Call":            "Call",
	"Truncate":        "Truncate",
}

// MarshalStatements serializes a slice of AST statements to JSON.
func MarshalStatements(stmts interface{}) ([]byte, error) {
	v := reflect.ValueOf(stmts)
	if v.Kind() == reflect.Slice {
		var result []interface{}
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			converted := valToJSON(reflect.ValueOf(item))
			converted = wrapStatement(item, converted)
			result = append(result, converted)
		}
		return json.MarshalIndent(result, "", "  ")
	}
	return json.MarshalIndent(valToJSON(reflect.ValueOf(stmts)), "", "  ")
}

func wrapStatement(stmt interface{}, converted interface{}) interface{} {
	rv := reflect.ValueOf(stmt)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	typeName := rv.Type().Name()
	if tag, ok := statementTypes[typeName]; ok {
		return map[string]interface{}{tag: converted}
	}
	return converted
}

func valToJSON(rv reflect.Value) interface{} {
	if !rv.IsValid() {
		return nil
	}
	if rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		return taggedUnion(rv.Elem())
	}
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		return valToJSON(rv.Elem())
	}
	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.CanInterface() {
			v := rv.Interface()
			if stringer, ok := v.(fmt.Stringer); ok {
				if s := stringer.String(); s != "" {
					return s
				}
			}
		}
		return rv.Int()
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	case reflect.Slice:
		if rv.IsNil() {
			return nil
		}
		l := rv.Len()
		if l == 0 {
			return []interface{}{}
		}
		result := make([]interface{}, l)
		for i := 0; i < l; i++ {
			result[i] = valToJSON(rv.Index(i))
		}
		return result
	case reflect.Map:
		if rv.IsNil() {
			return nil
		}
		result := make(map[string]interface{})
		for _, key := range rv.MapKeys() {
			if key.Kind() == reflect.String {
				result[key.String()] = valToJSON(rv.MapIndex(key))
			}
		}
		return result
	case reflect.Struct:
		return structToMap(rv)
	default:
		return fmt.Sprintf("%v", rv.Interface())
	}
}

func taggedUnion(rv reflect.Value) interface{} {
	if !rv.IsValid() {
		return nil
	}
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	}
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		return taggedUnion(rv.Elem())
	}
	if rv.Kind() == reflect.Slice {
		l := rv.Len()
		result := make([]interface{}, l)
		for i := 0; i < l; i++ {
			result[i] = valToJSON(rv.Index(i))
		}
		return result
	}
	if rv.Kind() == reflect.Struct {
		typeName := rv.Type().Name()
		fields := structToMap(rv)
		return map[string]interface{}{typeName: fields}
	}
	return valToJSON(rv)
}

// structToMap converts a struct to map[string]interface{} with snake_case keys.
// Handles both exported and unexported fields (via unsafe for the latter).
func structToMap(rv reflect.Value) map[string]interface{} {
	t := rv.Type()
	result := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := rv.Field(i)

		// Skip span fields
		if skipFields[f.Name] {
			continue
		}

		// Handle anonymous/embedded fields
		if f.Anonymous {
			if baseTypes[f.Name] || baseTypes[f.Type.Name()] {
				continue
			}
			// For embedded structs, merge their fields (skip span)
			embMap := structToMap(fv)
			for k, v := range embMap {
				if k == "span" || k == "span_val" {
					continue
				}
				result[k] = v
			}
			continue
		}

		// Get the value, handling unexported fields via unsafe access
		var val interface{}
		if f.IsExported() {
			val = valToJSON(fv)
		} else if !skipFields[f.Name] {
			// Use unsafe to access unexported fields (except span metadata)
			val = getUnexportedVal(fv)
		}

		if val != nil {
			jsonKey := toSnakeCase(f.Name)
			result[jsonKey] = val
		}
	}

	return result
}

// getUnexportedVal accesses unexported struct fields via unsafe pointer.
func getUnexportedVal(fv reflect.Value) interface{} {
	if !fv.IsValid() {
		return nil
	}

	// Create a new reflect.Value pointing to the same memory but accessible
	fv2 := reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
	return valToJSON(fv2)
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
