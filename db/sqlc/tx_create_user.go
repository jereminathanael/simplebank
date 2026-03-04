package db

import "context"


type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(user User) error
}

type CreateUserTxResult struct {
}

// CreateUserTx performs a money transfer from one account to the other.
func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		return err
	})

	return result, err
}