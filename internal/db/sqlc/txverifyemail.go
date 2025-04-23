package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type VerifyEmailTxParams struct {
	EmailId     int64
	SecreteCode string
}

type VerifyEmailTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

func (s *SQLStore) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.EmailId,
			SecretCode: arg.SecreteCode,
		})
		if err != nil {
			fmt.Println("oidoaidoai")
			return err
		}

		fmt.Println(result.VerifyEmail)

		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			Username: result.VerifyEmail.Username,
			IsEmailVerified: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		})

		return err
	})

	return result, err
}
