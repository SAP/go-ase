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

// TestASEType_GoType ensures that each ASEType returns a non-nil
// reflect.Type.
func TestASEType_GoType(t *testing.T) {
	for asetype, name := range type2string {
		t.Run(name,
			func(t *testing.T) {
				reflectType := asetype.GoType()
				if reflectType == nil {
					t.Errorf("%s.GoType() returned nil", name)
				}
			},
		)
	}
}
