package product

import "errors"

var (
	ErrUnknownProvider = errors.New("unknown payment provider")
	ErrNotPaid         = errors.New("payment not completed")
	ErrWrongAmount     = errors.New("amount paid does not match the order")
	ErrNoOrder         = errors.New("no such order for this account")
	ErrWrongCurrency   = errors.New("payment is not in the currency we priced")
)
