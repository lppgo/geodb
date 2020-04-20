package db

import (
	"github.com/dgraph-io/badger/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
)

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
