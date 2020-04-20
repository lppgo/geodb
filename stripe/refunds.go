package stripe

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/refund"
)

type RefundFunc func(params *stripe.RefundParams) (*stripe.RefundParams, error)

func RefundCharge(fn RefundFunc) error {
	params := &stripe.RefundParams{}
	params, err := fn(params)
	if err != nil {
		return err
	}
	_, err = refund.New(params)
	if err != nil {
		return err
	}
	return nil
}
