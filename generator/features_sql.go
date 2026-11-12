package generator

import (
	"fmt"
	"strings"

	"github.com/hushink/enumer/model"
)

type sqlFeature struct{}

func (sqlFeature) Name() string                   { return "sql" }
func (sqlFeature) Requires() []string             { return []string{"parse", "string"} }
func (sqlFeature) Validate(*model.EnumSpec) error { return nil }

func (sqlFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	b.WriteString("// Scan implements sql.Scanner. It does not validate that the value is declared.\n")
	fmt.Fprintf(&b, "func (x *%s) Scan(src any) error {\n", e.Name)
	fmt.Fprintf(&b, "\tif src == nil {\n\t\tvar zero %s\n\t\t*x = zero\n\t\treturn nil\n\t}\n", e.Name)
	b.WriteString("\tswitch v := src.(type) {\n")

	if e.IsString() {
		fmt.Fprintf(&b, "\tcase string:\n\t\t*x = %s(v)\n\t\treturn nil\n", e.Name)
		fmt.Fprintf(&b, "\tcase []byte:\n\t\t*x = %s(string(v))\n\t\treturn nil\n", e.Name)
	} else if isUnsigned(e) {
		b.WriteString("\tcase int64:\n")
		if e.Type.Bits < 64 {
			fmt.Fprintf(&b, "\t\tif v < 0 || uint64(v) > %d {\n", maxUnsigned(e.Type.Bits))
			fmt.Fprintf(&b, "\t\t\treturn fmt.Errorf(\"enumer: %%d is out of range for %s\", v)\n\t\t}\n", e.Type.Name)
		} else {
			b.WriteString("\t\tif v < 0 {\n")
			fmt.Fprintf(&b, "\t\t\treturn fmt.Errorf(\"enumer: %%d is out of range for %s\", v)\n\t\t}\n", e.Type.Name)
		}
		fmt.Fprintf(&b, "\t\t*x = %s(v)\n\t\treturn nil\n", e.Name)
		fmt.Fprintf(&b, "\tcase string:\n\t\tn, err := strconv.ParseUint(v, 10, %d)\n", e.Type.Bits)
		b.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		fmt.Fprintf(&b, "\t\t*x = %s(n)\n\t\treturn nil\n", e.Name)
		fmt.Fprintf(&b, "\tcase []byte:\n\t\tn, err := strconv.ParseUint(string(v), 10, %d)\n", e.Type.Bits)
		b.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		fmt.Fprintf(&b, "\t\t*x = %s(n)\n\t\treturn nil\n", e.Name)
	} else {
		b.WriteString("\tcase int64:\n")
		if e.Type.Bits < 64 {
			min, max := signedBounds(e.Type.Bits)
			fmt.Fprintf(&b, "\t\tif v < %d || v > %d {\n", min, max)
			fmt.Fprintf(&b, "\t\t\treturn fmt.Errorf(\"enumer: %%d is out of range for %s\", v)\n\t\t}\n", e.Type.Name)
		}
		fmt.Fprintf(&b, "\t\t*x = %s(v)\n\t\treturn nil\n", e.Name)
		fmt.Fprintf(&b, "\tcase string:\n\t\tn, err := strconv.ParseInt(v, 10, %d)\n", e.Type.Bits)
		b.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		fmt.Fprintf(&b, "\t\t*x = %s(n)\n\t\treturn nil\n", e.Name)
		fmt.Fprintf(&b, "\tcase []byte:\n\t\tn, err := strconv.ParseInt(string(v), 10, %d)\n", e.Type.Bits)
		b.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		fmt.Fprintf(&b, "\t\t*x = %s(n)\n\t\treturn nil\n", e.Name)
	}

	fmt.Fprintf(&b, "\tdefault:\n\t\treturn fmt.Errorf(\"enumer: cannot scan %%T into %s\", src)\n\t}\n}\n\n", e.Name)

	b.WriteString("// Value implements driver.Valuer. It does not validate that the value is declared.\n")
	fmt.Fprintf(&b, "func (x %s) Value() (driver.Value, error) {\n", e.Name)
	switch {
	case e.IsString():
		b.WriteString("\treturn string(x), nil\n")
	case isUnsigned(e) && e.Type.Bits == 64:
		b.WriteString("\tif uint64(x) > 9223372036854775807 {\n")
		fmt.Fprintf(&b, "\t\treturn nil, fmt.Errorf(\"enumer: %%d overflows int64\", uint64(x))\n\t}\n")
		b.WriteString("\treturn int64(x), nil\n")
	default:
		b.WriteString("\treturn int64(x), nil\n")
	}
	b.WriteString("}\n")

	imports := []model.Import{{Path: "database/sql/driver"}, {Path: "fmt"}}
	if !e.IsString() {
		imports = append(imports, model.Import{Path: "strconv"})
	}
	return Fragment{Code: b.String(), Imports: imports}, nil
}

type sqlIntFeature struct{}

func (sqlIntFeature) Name() string       { return "sql:int" }
func (sqlIntFeature) Requires() []string { return []string{"parse", "string"} }

func (sqlIntFeature) Validate(e *model.EnumSpec) error {
	if !e.IsString() {
		return fmt.Errorf("sql:int requires a string enum")
	}
	for i := range e.Values {
		if !e.Values[i].HasSQLNumber {
			return fmt.Errorf("value %q must declare number", e.Values[i].Name)
		}
	}
	return nil
}

func (sqlIntFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "var _%sByNumber = map[int64]%s{\n", e.Name, e.Name)
	seen := map[string]bool{}
	for i := range e.Values {
		v := &e.Values[i]
		if !v.HasSQLNumber || seen[v.SQLNumber] {
			continue
		}
		seen[v.SQLNumber] = true
		fmt.Fprintf(&b, "\t%s: %s,\n", v.SQLNumber, v.ConstName)
	}
	b.WriteString("}\n\n")

	b.WriteString("// Scan implements sql.Scanner.\n")
	fmt.Fprintf(&b, "func (x *%s) Scan(src any) error {\n", e.Name)
	fmt.Fprintf(&b, "\tif src == nil {\n\t\tvar zero %s\n\t\t*x = zero\n\t\treturn nil\n\t}\n", e.Name)
	b.WriteString("\tvar n int64\n")
	b.WriteString("\tswitch v := src.(type) {\n")
	b.WriteString("\tcase int64:\n\t\tn = v\n")
	b.WriteString("\tcase string:\n\t\tparsed, err := strconv.ParseInt(v, 10, 64)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tn = parsed\n")
	b.WriteString("\tcase []byte:\n\t\tparsed, err := strconv.ParseInt(string(v), 10, 64)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tn = parsed\n")
	fmt.Fprintf(&b, "\tdefault:\n\t\treturn fmt.Errorf(\"enumer: cannot scan %%T into %s\", src)\n\t}\n", e.Name)
	fmt.Fprintf(&b, "\tvalue, ok := _%sByNumber[n]\n", e.Name)
	fmt.Fprintf(&b, "\tif !ok {\n\t\treturn fmt.Errorf(\"enumer: %%d is not a valid %s\", n)\n\t}\n", e.Name)
	b.WriteString("\t*x = value\n\treturn nil\n}\n\n")

	b.WriteString("// Value implements driver.Valuer.\n")
	fmt.Fprintf(&b, "func (x %s) Value() (driver.Value, error) {\n", e.Name)
	b.WriteString("\tn, ok := x.sqlNumber()\n")
	fmt.Fprintf(&b, "\tif !ok {\n\t\treturn nil, fmt.Errorf(\"enumer: invalid %s value %%q\", string(x))\n\t}\n", e.Name)
	b.WriteString("\treturn n, nil\n}\n\n")

	fmt.Fprintf(&b, "func (x %s) sqlNumber() (int64, bool) {\n", e.Name)
	for i := range e.Values {
		v := &e.Values[i]
		if !v.HasSQLNumber {
			continue
		}
		fmt.Fprintf(&b, "\tif x == %s {\n\t\treturn %s, true\n\t}\n", v.ConstName, v.SQLNumber)
	}
	b.WriteString("\treturn 0, false\n}\n")

	return Fragment{
		Code:    b.String(),
		Imports: []model.Import{{Path: "database/sql/driver"}, {Path: "fmt"}, {Path: "strconv"}},
	}, nil
}

type sqlNullFeature struct{}

func (sqlNullFeature) Name() string       { return "sql:null" }
func (sqlNullFeature) Requires() []string { return nil }

func (sqlNullFeature) Validate(e *model.EnumSpec) error {
	hasSQL := e.HasFeature("sql")
	hasInt := e.HasFeature("sql:int")
	if hasSQL == hasInt {
		return fmt.Errorf("sql:null requires exactly one of sql or sql:int")
	}
	return nil
}

func (sqlNullFeature) Emit(e *model.EnumSpec) (Fragment, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "// Null%s represents a nullable %s.\n", e.Name, e.Name)
	fmt.Fprintf(&b, "type Null%s struct {\n\tEnum  %s\n\tValid bool\n}\n\n", e.Name, e.Name)

	fmt.Fprintf(&b, "// Scan implements sql.Scanner.\nfunc (n *Null%s) Scan(src any) error {\n", e.Name)
	b.WriteString("\tif src == nil {\n")
	fmt.Fprintf(&b, "\t\tvar zero %s\n", e.Name)
	b.WriteString("\t\tn.Enum = zero\n\t\tn.Valid = false\n\t\treturn nil\n\t}\n")
	b.WriteString("\tif err := n.Enum.Scan(src); err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tn.Valid = true\n\treturn nil\n}\n\n")

	b.WriteString("// Value implements driver.Valuer.\n")
	fmt.Fprintf(&b, "func (n Null%s) Value() (driver.Value, error) {\n", e.Name)
	b.WriteString("\tif !n.Valid {\n\t\treturn nil, nil\n\t}\n")
	b.WriteString("\treturn n.Enum.Value()\n}\n\n")

	b.WriteString("// MarshalJSON implements json.Marshaler.\n")
	fmt.Fprintf(&b, "func (n Null%s) MarshalJSON() ([]byte, error) {\n", e.Name)
	b.WriteString("\tif !n.Valid {\n\t\treturn []byte(\"null\"), nil\n\t}\n")
	b.WriteString("\treturn json.Marshal(n.Enum)\n}\n\n")

	b.WriteString("// UnmarshalJSON implements json.Unmarshaler.\n")
	fmt.Fprintf(&b, "func (n *Null%s) UnmarshalJSON(data []byte) error {\n", e.Name)
	b.WriteString("\tif string(data) == \"null\" {\n")
	fmt.Fprintf(&b, "\t\tvar zero %s\n", e.Name)
	b.WriteString("\t\tn.Enum = zero\n\t\tn.Valid = false\n\t\treturn nil\n\t}\n")
	fmt.Fprintf(&b, "\tvar v %s\n", e.Name)
	b.WriteString("\tif err := json.Unmarshal(data, &v); err != nil {\n\t\treturn err\n\t}\n")
	b.WriteString("\tn.Enum = v\n\tn.Valid = true\n\treturn nil\n}\n")

	return Fragment{
		Code:    b.String(),
		Imports: []model.Import{{Path: "database/sql/driver"}, {Path: "encoding/json"}},
	}, nil
}

func signedBounds(bits int) (int64, int64) {
	switch bits {
	case 8:
		return -128, 127
	case 16:
		return -32768, 32767
	case 32:
		return -2147483648, 2147483647
	default:
		return -9223372036854775808, 9223372036854775807
	}
}

func maxUnsigned(bits int) uint64 {
	switch bits {
	case 8:
		return 255
	case 16:
		return 65535
	case 32:
		return 4294967295
	default:
		return 18446744073709551615
	}
}
