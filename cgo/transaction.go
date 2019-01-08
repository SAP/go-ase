package cgo

// transaction is the struct which represents a database transaction.
type transaction struct {
	conn *C.CS_CONNECTION
}

func (transaction *transaction) Commit() error {
	// TODO
	return nil
}

func (transaction *transaction) Rollback() error {
	// TODO
	return nil
}
