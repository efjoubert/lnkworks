package embed

import (
	"io"
	"strings"
)

`

func TypeScriptJS() io.Reader {
	return strings.NewReader(strings.ReplaceAll(typescriptjs, "|'|", "`"))
}