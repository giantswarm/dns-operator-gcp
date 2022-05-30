package registrar

import (
	"errors"

	"google.golang.org/api/googleapi"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	RecordNS    = "NS"
	RecordA     = "A"
	RecordCNAME = "CNAME"
)

func hasHttpCode(err error, statusCode int) bool {
	var googleErr *googleapi.Error
	if errors.As(err, &googleErr) {
		if googleErr.Code == statusCode {
			return true
		}
	}

	return false
}
