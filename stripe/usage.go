package stripe

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/usagerecord"
)

type UsageFunc func(params *stripe.UsageRecordParams) (*stripe.UsageRecordParams, error)

func UpdateUsage(fn UsageFunc) error {
	params := &stripe.UsageRecordParams{}
	params, err := fn(params)
	if err != nil {
		return err
	}
	_, err = usagerecord.New(params)
	if err != nil {
		return err
	}
	return nil
}
