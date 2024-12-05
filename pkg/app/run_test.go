package app

import (
	"bytes"
	"testing"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/stretchr/testify/assert"
)

func Test_Run(t *testing.T) {
	t.Setenv("CREATE_FILES", "yes")

	buf := bytes.NewBuffer([]byte{})

	log.Default().SetOutput(buf)

	Run()

	assert.Contains(t, buf.String(), `{"level":"error","msg":"panic recover","panic":{"error":"failed to parse config: env: parse error on field \"CreateFilesByYourself\" of type \"bool\": strconv.ParseBool: parsing \"yes\": invalid syntax","stacktrace":`)
}
