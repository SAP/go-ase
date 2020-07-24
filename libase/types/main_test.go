package types

import "testing"

func TestType(t *testing.T) {
	typ := Type("MONEY")
	if typ != MONEY {
		t.Errorf("Type() returned wrong ASEType: %v", typ)
	}
}

func TestASEType_String(t *testing.T) {
	typ := UINT
	if typ.String() != "UINT" {
		t.Errorf("Received unexpected string: %s", typ.String())
	}
}

// nilTypes are types that haven't been implemented yet and as such are
// expected to return nil.
var nilTypes = map[ASEType]bool{
	ILLEGAL:        true,
	TEXTLOCATOR:    true,
	BOUNDARY:       true,
	IMAGELOCATOR:   true,
	UNITEXTLOCATOR: true,
	SENSITIVITY:    true,
	USER:           true,
	VOID:           true,
}

// TestASEType_GoReflectType ensures that each ASEType returns a non-nil
// reflect.Type.
func TestASEType_GoReflectType(t *testing.T) {
	for asetype, name := range type2string {
		t.Run(name,
			func(t *testing.T) {
				reflectType := asetype.GoType()

				shouldskip, ok := nilTypes[asetype]
				if ok && shouldskip {
					if reflectType != nil {
						t.Errorf("%s.GoReflectType() should return nil, returned: %v", name, reflectType)
					}
					return
				}

				if reflectType == nil {
					t.Errorf("%s.GoReflectType() returned nil", name)
				}
			},
		)
	}
}

func TestASEType_GoType(t *testing.T) {
	for asetype, name := range type2string {
		t.Run(name,
			func(t *testing.T) {
				goType := asetype.GoType()

				shouldskip, ok := nilTypes[asetype]
				if ok && shouldskip {
					if goType != nil {
						t.Errorf("%s.GoType() should return nil, returned: %v", name, goType)
					}
					return
				}

				if goType == nil {
					t.Errorf("%s.GoType() returned nil", name)
				}
			},
		)
	}
}
