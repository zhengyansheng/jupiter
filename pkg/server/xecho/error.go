// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xecho

import (
	"net/http"

	"github.com/zhengyansheng/jupiter/pkg/util/xerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errBadRequest         = xerror.InvalidArgument.WithMsg("bad request")
	errMicroInvoke        = xerror.Internal.WithMsg("invoke failed")
	errMicroInvokeLen     = xerror.Internal.WithMsg("invoke result not 2 item")
	errMicroInvokeInvalid = xerror.Internal.WithMsg("second invoke res not a error")
	errMicroResInvalid    = xerror.Internal.WithMsg("response is not valid")
)

// HTTPError wraps handler error.
type HTTPError struct {
	Code    int
	Message string
}

// NewHTTPError constructs a new HTTPError instance.
func NewHTTPError(code int, msg ...string) *HTTPError {
	he := &HTTPError{Code: code, Message: http.StatusText(code)}
	if len(msg) > 0 {
		he.Message = msg[0]
	}

	return he
}

// Errord return error message.
func (e HTTPError) Error() string {
	return e.Message
}

// ErrNotFound defines StatusNotFound error.
var ErrNotFound = HTTPError{
	Code:    http.StatusNotFound,
	Message: "not found",
}

var (
	// ErrGRPCResponseValid ...
	ErrGRPCResponseValid = status.Errorf(codes.Internal, "response valid")
	// ErrGRPCInvokeLen ...
	ErrGRPCInvokeLen = status.Errorf(codes.Internal, "invoke request without len 2 res")
)
