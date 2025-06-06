package pkg

import "strconv"

var _ error = (*Error)(nil)

type Error struct {
	Message string
	Code    int
}

func (e *Error) Error() string {
	if e.Code != 0 {
		return e.Message + " (code: " + strconv.Itoa(e.Code) + ")"
	}

	return e.Message
}
