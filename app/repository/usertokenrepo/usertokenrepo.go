package usertokenrepo

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

func (r *repositoryImpl) Get(ctx context.Context, userID string) (*oauth2.Token, error) {
	var accessToken, refreshToken string
	var expiry time.Time

	err := r.db.QueryRowContext(ctx, `
		SELECT access_token, refresh_token, token_expiry
		FROM user_tokens
		WHERE user_id = $1
	`, userID).Scan(&accessToken, &refreshToken, &expiry)
	if err != nil {
		return nil, fmt.Errorf("unable to get user token: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}, nil
}

func (r *repositoryImpl) Save(ctx context.Context, userID string, token *oauth2.Token) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_tokens (user_id, access_token, refresh_token, token_expiry)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			access_token  = $2,
			refresh_token = $3,
			token_expiry  = $4,
			updated_at    = now()
	`, userID, token.AccessToken, token.RefreshToken, token.Expiry)
	if err != nil {
		return fmt.Errorf("unable to save user token: %w", err)
	}
	return nil
}
