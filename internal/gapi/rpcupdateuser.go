package gapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/lucasHSantiago/gobank/internal/db/sqlc"
	"github.com/lucasHSantiago/gobank/internal/db/util"
	"github.com/lucasHSantiago/gobank/internal/validator"
	"github.com/lucasHSantiago/gobank/proto/gen"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *gen.UpdateUserRequest) (*gen.UpdateUserResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violations := validUpdateUserRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
		}

		arg.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}

		arg.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	response := &gen.UpdateUserResponse{
		User: convertUser(user),
	}

	return response, nil
}

func validUpdateUserRequest(req *gen.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if req.FullName != nil {
		if err := validator.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if req.Password != nil {
		if err := validator.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}

	if req.Email != nil {
		if err := validator.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return violations
}
