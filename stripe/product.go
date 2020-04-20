package stripe

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/product"
)

type ProductFunc func(params *stripe.ProductParams) (*stripe.ProductParams, error)

func NewProduct(fn ProductFunc) (*stripe.Product, error) {
	params := &stripe.ProductParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return product.New(params)
}

func GetProduct(id string, fn ProductFunc) (*stripe.Product, error) {
	params := &stripe.ProductParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return product.Get(id, params)
}

func UpdateProduct(id string, fn ProductFunc) (*stripe.Product, error) {
	params := &stripe.ProductParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return product.Update(id, params)
}

type ProductListFunc func(params *stripe.ProductListParams) (*stripe.ProductListParams, error)

func GetProducts(id string, fn ProductListFunc) (*product.Iter, error) {
	params := &stripe.ProductListParams{}
	params, err := fn(params)
	if err != nil {
		return nil, err
	}
	return product.List(params), nil
}
