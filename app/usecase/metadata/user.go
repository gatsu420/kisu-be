package metadata

import (
	"context"

	"github.com/gatsu420/kisu-be/app/repository/pgrepo"
	"golang.org/x/oauth2"
)

type InsertUserArgs struct {
	Email string
}

type InsertUserResult struct {
	UserID string
}

func (u *usecaseImpl) InsertUser(ctx context.Context, args InsertUserArgs) (InsertUserResult, error) {
	result, err := u.pgRepo.InsertUser(ctx, pgrepo.InsertUserArgs{
		Email: args.Email,
	})
	if err != nil {
		return InsertUserResult{}, err
	}

	return InsertUserResult{
		UserID: result.UserID,
	}, nil
}

type InsertUserTokenArgs struct {
	UserID string
	Token  *oauth2.Token
}

func (u *usecaseImpl) InsertUserToken(ctx context.Context, args InsertUserTokenArgs) error {
	err := u.pgRepo.InsertUserToken(ctx, pgrepo.InsertUserTokenArgs{
		UserID: args.UserID,
		Token:  args.Token,
	})
	if err != nil {
		return err
	}

	return nil
}

type GetUserTokenArgs struct {
	UserID string
}

type GetUserTokenResult struct {
	Token *oauth2.Token
}

func (u *usecaseImpl) GetUserToken(ctx context.Context, args GetUserTokenArgs) (GetUserTokenResult, error) {
	result, err := u.pgRepo.GetUserToken(ctx, pgrepo.GetUserTokenArgs{
		UserID: args.UserID,
	})
	if err != nil {
		return GetUserTokenResult{}, err
	}

	return GetUserTokenResult{
		Token: result.Token,
	}, nil
}
