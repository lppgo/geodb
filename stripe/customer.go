package stripe

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

type CustomerFunc func(params *stripe.CustomerParams) (*stripe.CustomerParams, error)

func NewCustomer(fn CustomerFunc) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return customer.New(params)
}

func GetCustomer(id string, fn CustomerFunc) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return customer.Get(id, params)
}

func UpdateCustomer(id string, fn CustomerFunc) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return customer.Update(id, params)
}

type CustomerListFunc func(params *stripe.CustomerListParams) (*stripe.CustomerListParams, error)

func ListCustomers(id string, fn CustomerListFunc) (*customer.Iter, error) {
	params := &stripe.CustomerListParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return customer.List(params), nil
}
