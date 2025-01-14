package app_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg/app"
)

func Test_Run(t *testing.T) {
	t.Setenv("CREATE_FILES", "yes")

	buf := bytes.NewBuffer([]byte{})

	log.Default().SetOutput(buf)

	app.Run()

	assert.Contains(t, buf.String(), `{"level":"error","msg":"panic recover","panic":{"error":"failed to parse config: env: parse error on field \"CreateFilesByYourself\" of type \"bool\": strconv.ParseBool: parsing \"yes\": invalid syntax","stacktrace":`)
}
