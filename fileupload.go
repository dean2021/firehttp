package firehttp

import (
	"io"
	"strings"
)

type FileUpload struct {
	FileName string

	FileBody io.ReadCloser

	FieldName string

	// 默认 application/octet-stream
	FileMime string
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
