package db

import (
	"github.com/autom8ter/userdb/auth"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/autom8ter/userdb/stripe"
	"github.com/dgraph-io/badger/v2"
	"github.com/gogo/protobuf/proto"
	stripe2 "github.com/stripe/stripe-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
	"time"
)

func Login(db *badger.DB, usr *auth.GoogleUser) (*api.UserDetail, error) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	item, err := txn.Get([]byte(usr.Email))
	if err == nil && item != nil && item.UserMeta() == 1 {
		res, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}
		var detail = &api.UserDetail{}
		if err := proto.Unmarshal(res, detail); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to unmarshal protobuf: %s", err.Error())
		}
		return detail, nil
	}
	detail := &api.UserDetail{
		Email:       usr.Email,
		Name:        usr.Name,
		UpdatedUnix: time.Now().Unix(),
	}
	bits, err := proto.Marshal(detail)
	if err != nil {
		return nil, err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(detail.Email),
		Value:    bits,
		UserMeta: 1,
	}); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return detail, nil
}

func Set(db *badger.DB, obj *api.UserDetail) (*api.UserDetail, error) {
	if err := obj.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if obj.UpdatedUnix == 0 {
		obj.UpdatedUnix = time.Now().Unix()
	}
	detail := obj
	bits, err := proto.Marshal(detail)
	if err != nil {
		return nil, err
	}
	txn := db.NewTransaction(true)
	defer txn.Discard()
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(obj.Email),
		Value:    bits,
		UserMeta: 1,
	}); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return detail, nil
}

func Get(db *badger.DB, emails []string) (map[string]*api.UserDetail, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	objects := map[string]*api.UserDetail{}
	if len(emails) == 0 {
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
				var obj = &api.UserDetail{}
				if err := proto.Unmarshal(res, obj); err != nil {
					return nil, status.Errorf(codes.Internal, "(keys) %s failed to unmarshal protobuf: %s", string(item.Key()), err.Error())
				}
				objects[string(item.Key())] = obj
			}
		}
	} else {
		for _, key := range emails {
			i, err := txn.Get([]byte(key))
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
			}
			if i.UserMeta() != 1 {
				continue
			}
			res, err := i.ValueCopy(nil)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
			}
			var obj = &api.UserDetail{}
			if err := proto.Unmarshal(res, obj); err != nil {
				return nil, status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
			}
			objects[key] = obj
		}
	}
	return objects, nil
}

func GetRegex(db *badger.DB, regex string) (map[string]*api.UserDetail, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	objects := map[string]*api.UserDetail{}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iter := txn.NewIterator(opts)
	defer iter.Close()
	for iter.Rewind(); iter.Valid(); iter.Next() {
		item := iter.Item()
		if item.UserMeta() != 1 {
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
			var obj = &api.UserDetail{}
			if err := proto.Unmarshal(res, obj); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to unmarshal protobuf: %s", err.Error())
			}
			objects[string(item.Key())] = obj
		}

	}
	return objects, nil
}

func Delete(db *badger.DB, keys []string) error {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	if len(keys) > 0 && keys[0] == "*" {
		if err := db.DropAll(); err != nil {
			return status.Errorf(codes.Internal, "failed to delete key: %s", err.Error())
		}
	} else {
		for _, key := range keys {
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

func GetEmails(db *badger.DB) []string {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	keys := []string{}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iter := txn.NewIterator(opts)
	for iter.Rewind(); iter.Valid(); iter.Next() {
		item := iter.Item()
		if item.UserMeta() != 1 {
			continue
		}
		keys = append(keys, string(item.Key()))
	}
	iter.Close()
	return keys
}

func GetRegexEmails(db *badger.DB, regex string) ([]string, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	keys := []string{}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iter := txn.NewIterator(opts)
	for iter.Rewind(); iter.Valid(); iter.Next() {
		item := iter.Item()
		if item.UserMeta() != 1 {
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

func AddPlan(db *badger.DB, email, plan string) (*api.UserDetail, error) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	i, err := txn.Get([]byte(email))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
	}
	if i.UserMeta() != 1 {
		return nil, err
	}
	res, err := i.ValueCopy(nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
	}
	var detail = &api.UserDetail{}
	if err := proto.Unmarshal(res, detail); err != nil {
		return nil, status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
	}
	cust, err := stripe.UpdateCustomer(detail.GetPayment().GetCustomerId(), func(params *stripe2.CustomerParams) (params2 *stripe2.CustomerParams, err error) {
		params.Plan = stripe2.String(plan)
		return params, nil
	})
	if err != nil {
		return nil, err
	}
	if detail.Payment == nil {
		detail.Payment = &api.Payment{
			CustomerId: cust.ID,
		}
	}
	for _, sub := range cust.Subscriptions.Data {
		detail.Payment.Subscriptions = append(detail.Payment.Subscriptions, &api.Subscription{
			Subscription: sub.ID,
			Plan:         sub.Plan.ID,
			Product:      sub.Plan.Product.ID,
		})
	}

	bits, err := proto.Marshal(detail)
	if err != nil {
		return nil, err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(detail.GetEmail()),
		Value:    bits,
		UserMeta: 1,
	}); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return detail, nil
}

func SetSource(db *badger.DB, email, source string) (*api.UserDetail, error) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	i, err := txn.Get([]byte(email))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
	}
	if i.UserMeta() != 1 {
		return nil, err
	}
	res, err := i.ValueCopy(nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
	}
	var detail = &api.UserDetail{}
	if err := proto.Unmarshal(res, detail); err != nil {
		return nil, status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
	}
	_, err = stripe.UpdateCustomer(detail.GetPayment().GetCustomerId(), func(params *stripe2.CustomerParams) (params2 *stripe2.CustomerParams, err error) {
		params.Source = &stripe2.SourceParams{
			Token: &source,
		}
		return params, nil
	})
	detail.Payment.HasSource = true
	if err != nil {
		return nil, err
	}
	bits, err := proto.Marshal(detail)
	if err != nil {
		return nil, err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(detail.Email),
		Value:    bits,
		UserMeta: 1,
	}); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return detail, nil
}

func IncPlanUsage(db *badger.DB, increment int64, email, plan string) error {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	i, err := txn.Get([]byte(email))
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
	}
	if i.UserMeta() != 1 {
		return err
	}
	res, err := i.ValueCopy(nil)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
	}
	var detail = &api.UserDetail{}
	if err := proto.Unmarshal(res, detail); err != nil {
		return status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
	}
	if detail.Payment == nil || detail.Payment.CustomerId == "" || !detail.Payment.HasSource {
		return status.Errorf(codes.FailedPrecondition, "user %s does not have a payment method", detail.Email)
	}
	var item string
	for _, sub := range detail.GetPayment().GetSubscriptions() {
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
	bits, err := proto.Marshal(detail)
	if err != nil {
		return err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(detail.Email),
		Value:    bits,
		UserMeta: 1,
	}); err != nil {
		return err
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

func Charge(db *badger.DB, amount int64, accountName, description string, meta map[string]string) (string, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	i, err := txn.Get([]byte(accountName))
	if err != nil {
		return "", status.Errorf(codes.InvalidArgument, "failed to get key: %s", err.Error())
	}
	if i.UserMeta() != 1 {
		return "", err
	}
	res, err := i.ValueCopy(nil)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to copy data: %s", err.Error())
	}
	var detail = &api.UserDetail{}
	if err := proto.Unmarshal(res, detail); err != nil {
		return "", status.Errorf(codes.Internal, "(all) failed to unmarshal protobuf: %s", err.Error())
	}
	if detail.Payment == nil || detail.Payment.CustomerId == "" || !detail.Payment.HasSource {
		return "", status.Errorf(codes.FailedPrecondition, "account %s does not have a payment method", detail.Name)
	}
	charge, err := stripe.NewCharge(func(params *stripe2.ChargeParams) (params2 *stripe2.ChargeParams, err error) {
		params.Amount = &amount
		params.Currency = stripe2.String(string(stripe2.CurrencyUSD))
		params.Customer = &detail.Payment.CustomerId
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

func UpdateCharge(chargeID, description string, amount int64, meta map[string]string) (string, error) {
	charge, err := stripe.UpdateCharge(chargeID, func(params *stripe2.ChargeParams) (params2 *stripe2.ChargeParams, err error) {
		params.Amount = &amount
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

func RefundCharge(chargeID string) (bool, error) {
	if err := stripe.RefundCharge(func(params *stripe2.RefundParams) (params2 *stripe2.RefundParams, err error) {
		params.Charge = &chargeID
		return params, nil
	}); err != nil {
		return false, err
	}
	return true, nil
}
