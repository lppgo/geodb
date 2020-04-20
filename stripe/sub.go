package stripe

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
)

type SubFunc func(params *stripe.SubscriptionParams) (*stripe.SubscriptionParams, error)

func NewSubscription(fn SubFunc) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return sub.New(params)
}

func GetSubscription(id string, fn SubFunc) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return sub.Get(id, params)
}

func UpdateSubscription(id string, fn SubFunc) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return sub.Update(id, params)
}

type SubListFunc func(params *stripe.SubscriptionListParams) (*stripe.SubscriptionListParams, error)

func ListSubscriptions(fn SubListFunc) (*sub.Iter, error) {
	params := &stripe.SubscriptionListParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return sub.List(params), nil
}
