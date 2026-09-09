package subscription

import (
	"errors"

	"github.com/plutov/paypal/v4"
	"github.com/stripe/stripe-go/v81"
	"github.com/syncloud/redirect/model"
)

func serviceError(err error) error {
	if err == nil {
		return nil
	}
	var pp *paypal.ErrorResponse
	if errors.As(err, &pp) {
		code := 400
		if pp.Response != nil {
			code = pp.Response.StatusCode
		}
		message := pp.Message
		if len(pp.Details) > 0 && pp.Details[0].Description != "" {
			message = pp.Details[0].Description
		}
		if message == "" {
			message = "The payment provider rejected the request."
		}
		return model.NewServiceErrorWithCode(message, code)
	}
	var se *stripe.Error
	if errors.As(err, &se) {
		code := 400
		if se.HTTPStatusCode != 0 {
			code = se.HTTPStatusCode
		}
		message := se.Msg
		if message == "" {
			message = "The payment provider rejected the request."
		}
		return model.NewServiceErrorWithCode(message, code)
	}
	return err
}
