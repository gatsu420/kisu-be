package userrepo

import (
	"context"
	"fmt"
)

func (r *repositoryImpl) UpsertByEmail(ctx context.Context, email string) (string, error) {
	var userID string

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (email)
		VALUES ($1)
		ON CONFLICT (email) DO UPDATE SET email = $1
		RETURNING id
	`, email).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("unable to upsert user: %w", err)
	}

	return userID, nil
}
