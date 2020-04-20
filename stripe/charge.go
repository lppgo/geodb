package stripe

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
)

type ChargeFunc func(params *stripe.ChargeParams) (*stripe.ChargeParams, error)

func NewCharge(fn ChargeFunc) (*stripe.Charge, error) {
	params := &stripe.ChargeParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return charge.New(params)
}

func GetCharge(id string, fn ChargeFunc) (*stripe.Charge, error) {
	params := &stripe.ChargeParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return charge.Get(id, params)
}

func UpdateCharge(id string, fn ChargeFunc) (*stripe.Charge, error) {
	params := &stripe.ChargeParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return charge.Update(id, params)
}

type ChargeListFunc func(params *stripe.ChargeListParams) (*stripe.ChargeListParams, error)

func GetCharges(id string, fn ChargeListFunc) (*charge.Iter, error) {
	params := &stripe.ChargeListParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return charge.List(params), nil
}
