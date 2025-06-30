package errors

import "fmt"

type MyError struct {
	Op  string
	Err error
}

func (e *MyError) Error() string {
	return fmt.Sprintf("operation %s failed: %v", e.Op, e.Err)
}

func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &MyError{Op: op, Err: err}
}
