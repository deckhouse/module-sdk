package pkg

import (
	"io"
)

type Outputer interface {
	WriteOutput(writer io.Writer) error
}
