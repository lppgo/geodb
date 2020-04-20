package db

import (
	"fmt"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/dgraph-io/badger/v2"
	"github.com/golang/protobuf/proto"
)

func SetUserAccount(db *badger.DB, userEmail string, accountName string) (*api.UserDetail, error) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	acc := &api.Account{}
	usr := &api.UserDetail{}
	{
		item, err := txn.Get([]byte(userEmail))
		if err != nil {
			return nil, err
		}
		if item.UserMeta() != 1 {
			return nil, fmt.Errorf("%s is not a user. metadata: %s", userEmail, string(item.UserMeta()))
		}
		res, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}
		if err := proto.Unmarshal(res, usr); err == nil {
			return nil, err
		}
	}
	{
		item, err := txn.Get([]byte(accountName))
		if err != nil {
			return nil, err
		}
		if item.UserMeta() != 2 {
			return nil, fmt.Errorf("%s is not a account. metadata: %s", accountName, string(item.UserMeta()))
		}
		res, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}
		if err := proto.Unmarshal(res, acc); err == nil {
			return nil, err
		}
	}
	usr.Account = acc
	bits, err := proto.Marshal(usr)
	if err != nil {
		return nil, err
	}
	if err := txn.SetEntry(&badger.Entry{
		Key:      []byte(userEmail),
		Value:    bits,
		UserMeta: 1,
	}); err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return usr, nil
}
