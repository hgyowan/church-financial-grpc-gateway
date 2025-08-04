package middleware

import (
	"context"
	pkgError "github.com/hgyowan/go-pkg-library/error"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/proto"
)

func ResponseEnvelope(ctx context.Context, response proto.Message) (interface{}, error) {
	switch response.(type) {
	case *httpbody.HttpBody:
		return response, nil
	}
	return map[string]any{
		"code": pkgError.StatusOk,
		"data": response,
	}, nil
}
