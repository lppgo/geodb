package stripe

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
)

type PlanFunc func(params *stripe.PlanParams) (*stripe.PlanParams, error)

func NewPlan(fn PlanFunc) (*stripe.Plan, error) {
	params := &stripe.PlanParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return plan.New(params)
}

func GetPlan(id string, fn PlanFunc) (*stripe.Plan, error) {
	params := &stripe.PlanParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return plan.Get(id, params)
}

func UpdatePlan(id string, fn PlanFunc) (*stripe.Plan, error) {
	params := &stripe.PlanParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return plan.Update(id, params)
}

type PlanListFunc func(params *stripe.PlanListParams) (*stripe.PlanListParams, error)

func GetPlans(fn PlanListFunc) (*plan.Iter, error) {
	params := &stripe.PlanListParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return plan.List(params), nil
}
