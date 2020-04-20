package db

import (
	"github.com/autom8ter/userdb/config"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/autom8ter/userdb/stripe"
	"github.com/dgraph-io/badger/v2"
	"github.com/golang/protobuf/proto"
	stripe2 "github.com/stripe/stripe-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
	"time"
)

func NewAccount(db *badger.DB, name, adminEmail string, meta map[string]string) (*api.Account, error) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	acc := &api.Account{
		Name:        name,
		AdminEmail:  adminEmail,
		Metadata:    meta,
		UpdatedUnix: time.Now().Unix(),
	}
	cust, err := stripe.NewCustomer(func(params *stripe2.CustomerParams) (params2 *stripe2.CustomerParams, err error) {
		params.Name = &acc.Name
		params.Email = &acc.AdminEmail
		params.Metadata = acc.Metadata
		params.Plan = stripe2.String(config.Config.GetString("USERDB_DEFAULT_ACCOUNT_PLAN"))
		return params, nil
	})
	if err != nil {
		return nil, err
	}
	acc.Payment = &api.Payment{
		CustomerId: cust.ID,
	}
	for _, sub := range cust.Subscriptions.Data {
		acc.Payment.Subscriptions = append(acc.Payment.Subscriptions, &api.Subscription{
			Subscription: sub.ID,
			Plan:         sub.Plan.ID,
			Product:      sub.Plan.Product.ID,
		})
	}
	bits, err := proto.Marshal(acc)
	if err != nil {
		return nil, err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(acc.Name),
		Value:    bits,
		UserMeta: 2,
	}); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return acc, nil
}

func GetAccount(db *badger.DB, names []string) (map[string]*api.Account, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	objects := map[string]*api.Account{}
	if len(names) == 0 {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			if item.UserMeta() != 1 {
				continue
			}
			res, err := item.ValueCopy(nil)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
			}
			if len(res) > 0 {
				var obj = &api.Account{}
				if err := proto.Unmarshal(res, obj); err != nil {
					return nil, status.Errorf(codes.Internal, "(keys) %s failed to unmarshal protobuf: %s", string(item.Key()), err.Error())
				}
				objects[string(item.Key())] = obj
			}
		}
	} else {
		for _, key := range names {
			i, err := txn.Get([]byte(key))
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
			}
			if i.UserMeta() != 2 {
				continue
			}
			res, err := i.ValueCopy(nil)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
			}
			var obj = &api.Account{}
			if err := proto.Unmarshal(res, obj); err != nil {
				return nil, status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
			}
			objects[key] = obj
		}
	}
	return objects, nil
}

func GetAccountRegex(db *badger.DB, regex string) (map[string]*api.Account, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	objects := map[string]*api.Account{}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iter := txn.NewIterator(opts)
	defer iter.Close()
	for iter.Rewind(); iter.Valid(); iter.Next() {
		item := iter.Item()
		if item.UserMeta() != 2 {
			continue
		}
		match, err := regexp.MatchString(regex, string(item.Key()))
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to match regex: %s", err.Error())
		}
		if match {
			res, err := item.ValueCopy(nil)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
			}
			var obj = &api.Account{}
			if err := proto.Unmarshal(res, obj); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to unmarshal protobuf: %s", err.Error())
			}
			objects[string(item.Key())] = obj
		}

	}
	return objects, nil
}

func DeleteAccount(db *badger.DB, names []string) error {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	if len(names) > 0 && names[0] == "*" {
		if err := db.DropAll(); err != nil {
			return status.Errorf(codes.Internal, "failed to delete key: %s", err.Error())
		}
	} else {
		for _, key := range names {
			if err := txn.Delete([]byte(key)); err != nil {
				return status.Errorf(codes.Internal, "failed to delete key: %s %s", key, err.Error())
			}
		}
	}
	if err := txn.Commit(); err != nil {
		return status.Errorf(codes.Internal, "failed to delete keys %s", err.Error())
	}
	return nil
}

func AddAccountPlan(db *badger.DB, accountName, plan string) (*api.Account, error) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	i, err := txn.Get([]byte(accountName))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
	}
	if i.UserMeta() != 2 {
		return nil, err
	}
	res, err := i.ValueCopy(nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
	}
	var acc = &api.Account{}
	if err := proto.Unmarshal(res, acc); err != nil {
		return nil, status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
	}
	cust, err := stripe.UpdateCustomer(acc.GetPayment().GetCustomerId(), func(params *stripe2.CustomerParams) (params2 *stripe2.CustomerParams, err error) {
		params.Plan = stripe2.String(plan)
		return params, nil
	})
	if err != nil {
		return nil, err
	}
	if acc.Payment == nil {
		acc.Payment = &api.Payment{
			CustomerId: cust.ID,
		}
	}
	for _, sub := range cust.Subscriptions.Data {
		acc.Payment.Subscriptions = append(acc.Payment.Subscriptions, &api.Subscription{
			Subscription: sub.ID,
			Plan:         sub.Plan.ID,
			Product:      sub.Plan.Product.ID,
		})
	}

	bits, err := proto.Marshal(acc)
	if err != nil {
		return nil, err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(acc.Name),
		Value:    bits,
		UserMeta: 2,
	}); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return acc, nil
}

func SetAccountSource(db *badger.DB, accountName, source string) (*api.Account, error) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	i, err := txn.Get([]byte(accountName))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
	}
	if i.UserMeta() != 2 {
		return nil, err
	}
	res, err := i.ValueCopy(nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
	}
	var acc = &api.Account{}
	if err := proto.Unmarshal(res, acc); err != nil {
		return nil, status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
	}
	_, err = stripe.UpdateCustomer(acc.GetPayment().GetCustomerId(), func(params *stripe2.CustomerParams) (params2 *stripe2.CustomerParams, err error) {
		params.Source = &stripe2.SourceParams{
			Token: &source,
		}
		return params, nil
	})
	if err != nil {
		return nil, err
	}
	bits, err := proto.Marshal(acc)
	if err != nil {
		return nil, err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(acc.Name),
		Value:    bits,
		UserMeta: 2,
	}); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return acc, nil
}

func GetAccountNames(db *badger.DB) ([]string, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	keys := []string{}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iter := txn.NewIterator(opts)
	for iter.Rewind(); iter.Valid(); iter.Next() {
		item := iter.Item()
		if item.UserMeta() != 2 {
			continue
		}
		keys = append(keys, string(item.Key()))
	}
	iter.Close()
	return keys, nil
}

func GetAccountNamesRegex(db *badger.DB, regex string) ([]string, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	keys := []string{}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iter := txn.NewIterator(opts)
	for iter.Rewind(); iter.Valid(); iter.Next() {
		item := iter.Item()
		if item.UserMeta() != 2 {
			continue
		}
		match, err := regexp.MatchString(string(regex), string(item.Key()))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if match {
			keys = append(keys, string(item.Key()))
		}
	}
	iter.Close()
	return keys, nil
}

func IncAccountPlanUsage(db *badger.DB, increment int64, accountName, plan string) error {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	i, err := txn.Get([]byte(accountName))
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
	}
	if i.UserMeta() != 2 {
		return err
	}
	res, err := i.ValueCopy(nil)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
	}
	var acc = &api.Account{}
	if err := proto.Unmarshal(res, acc); err != nil {
		return status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
	}
	if acc.Payment == nil || acc.Payment.CustomerId == "" || !acc.Payment.HasSource {
		return status.Errorf(codes.FailedPrecondition, "account %s does not have a payment method", acc.Name)
	}
	var item string
	for _, sub := range acc.Payment.Subscriptions {
		if sub.Plan != "" && sub.Plan == plan {
			item = sub.Item
			break
		}
	}
	if item == "" {
		return status.Errorf(codes.InvalidArgument, "user does not have plan: %s", plan)
	}
	err = stripe.UpdateUsage(func(params *stripe2.UsageRecordParams) (params2 *stripe2.UsageRecordParams, err error) {
		params.SubscriptionItem = &item
		params.Quantity = &increment
		params.Timestamp = stripe2.Int64(time.Now().Unix())
		return params, nil
	})
	if err != nil {
		return err
	}
	bits, err := proto.Marshal(acc)
	if err != nil {
		return err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(acc.Name),
		Value:    bits,
		UserMeta: 2,
	}); err != nil {
		return err
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

func ChargeAccount(db *badger.DB, amount int64, accountName, description string, meta map[string]string) (string, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	i, err := txn.Get([]byte(accountName))
	if err != nil {
		return "", status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
	}
	if i.UserMeta() != 2 {
		return "", err
	}
	res, err := i.ValueCopy(nil)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
	}
	var acc = &api.Account{}
	if err := proto.Unmarshal(res, acc); err != nil {
		return "", status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
	}
	if acc.Payment == nil || acc.Payment.CustomerId == "" || !acc.Payment.HasSource {
		return "", status.Errorf(codes.FailedPrecondition, "account %s does not have a payment method", acc.Name)
	}
	charge, err := stripe.NewCharge(func(params *stripe2.ChargeParams) (params2 *stripe2.ChargeParams, err error) {
		params.Amount = &amount
		params.Currency = stripe2.String(string(stripe2.CurrencyUSD))
		params.Customer = &acc.Payment.CustomerId
		params.Metadata = meta
		if description != "" {
			params.Description = &description
		}
		return params, nil
	})
	if err != nil {
		return "", err
	}
	return charge.ID, nil
}
