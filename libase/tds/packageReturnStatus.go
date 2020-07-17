package tds

import "fmt"

var _ Package = (*ReturnStatusPackage)(nil)

type ReturnStatusPackage struct {
	returnValue int32
}

func (pkg *ReturnStatusPackage) ReadFrom(ch BytesChannel) error {
	var err error

	pkg.returnValue, err = ch.Int32()
	if err != nil {
		return fmt.Errorf("error reading return value: %w", err)
	}

	return nil
}

func (pkg ReturnStatusPackage) WriteTo(ch BytesChannel) error {
	err := ch.WriteInt32(pkg.returnValue)
	if err != nil {
		return fmt.Errorf("error writing return value: %w", err)
	}

	return nil
}

func (pkg ReturnStatusPackage) ReturnValue() int {
	return int(pkg.returnValue)
}

func (pkg ReturnStatusPackage) String() string {
	return fmt.Sprintf("%T(%d)", pkg, pkg.returnValue)
}
