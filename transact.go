package main

import (
	"context"

	"github.com/mjl-/bstore"
)

func _dbread(ctx context.Context, fn func(tx *bstore.Tx)) {
	err := database.Read(ctx, func(tx *bstore.Tx) error {
		fn(tx)
		return nil
	})
	_checkf(err, "transaction")
}

func _dbwrite(ctx context.Context, fn func(tx *bstore.Tx)) {
	err := database.Write(ctx, func(tx *bstore.Tx) error {
		fn(tx)
		return nil
	})
	_checkf(err, "transaction")
}
