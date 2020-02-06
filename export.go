package spider

import (
	"bytes"
	"net/http"
)

func DefaultExpandPolicy(req *http.Request, body *bytes.Buffer) URL {
	return defaultExpandPolicy(req, body)
}

func AppendSlice(m []string, check func(ele string) bool, arg ...string) []string {
	return appendSlice(m, check, arg...)
}
