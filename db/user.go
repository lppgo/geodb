package db

import (
	"github.com/autom8ter/userdb/auth"
	api "github.com/autom8ter/userdb/gen/go/userdb"
	"github.com/dgraph-io/badger/v2"
	"github.com/gogo/protobuf/proto"
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

func Get(db *badger.DB, keys []string) (map[string]*api.UserDetail, error) {
	txn := db.NewTransaction(false)
	defer txn.Discard()
	objects := map[string]*api.UserDetail{}
	if len(keys) == 0 {
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
		for _, key := range keys {
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
