package capability

import "fmt"

func ExampleNewCapability() {
	caps := []*Capability{
		NewCapability("cap1", "0.2.5-6", "0.4"),
		NewCapability("cap2", "1.0.2-alpha"),
		NewCapability("cap3"),
		NewCapability("cap4", "1.0.0", "", "0.2.5", "0.2.6"),
	}

	for _, cap := range caps {
		fmt.Println(cap)
	}
	// Output:
	// Capability cap1 -> ('0.2.5-6' -> '0.4')
	// Capability cap2 -> ('1.0.2-alpha' -> '')
	// Capability cap3 -> ()
	// Capability cap4 -> ('1.0.0' -> '', '0.2.5' -> '0.2.6')
}
