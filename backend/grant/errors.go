package grant

import "errors"

var ErrNoCode = errors.New("that code is not one of ours")
var ErrUsedUp = errors.New("that code has already been used")
