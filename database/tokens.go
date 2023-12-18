package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zeebo/errs"

	"kitchen_nerd/tokens"
)

// ensures that usersDB implements users.DB.
var _ tokens.DB = (*tokensDB)(nil)

var ErrTokens = errs.Class("tokens repository error")

const (
	createToken = ``
)

// tokensDB provides access to users db.
//
// architecture: Database
type tokensDB struct {
	pool *pgxpool.Pool
}

// GetToken returns user's token from the database.
func (tokensDB *tokensDB) GetToken(
	ctx context.Context, token string,
) (tokens.UserToken, error) {
	fields := map[string]interface{}{
		"token": token,
	}
	return tokensDB.GetTokenByFields(ctx, fields)
}

func (tokensDB *tokensDB) GetTokenByFields(
	ctx context.Context, fields map[string]interface{},
) (tokens.UserToken, error) {
	query := `SELECT id, user_id, token, expired_at,  created_at
			  FROM users_tokens`

	qParams := make([]interface{}, 0, len(fields))
	if len(fields) > 0 {
		query += " WHERE "

		i := 1
		and := ""
		for field, value := range fields {
			query += fmt.Sprintf("%s (%s = $%d)", and, field, i)
			qParams = append(qParams, value)
			and = " AND "
			i++
		}
	}

	var userToken tokens.UserToken
	row := tokensDB.pool.QueryRow(ctx, query, qParams...)

	err := row.Scan(
		&userToken.ID,
		&userToken.UserID,
		&userToken.Token,
		&userToken.ExpiredAt,
		&userToken.CreatedAt)
	if err != nil {
		if errs.Is(err, pgx.ErrNoRows) {
			return userToken, tokens.ErrNoToken.Wrap(err)
		}

		return userToken, ErrTokens.Wrap(err)
	}

	return userToken, nil
}

// GetTokenByID returns user's token from the database.
func (tokensDB *tokensDB) GetTokenByID(
	ctx context.Context, id uuid.UUID,
) (tokens.UserToken, error) {
	fields := map[string]interface{}{
		"user_id": id,
	}
	return tokensDB.GetTokenByFields(ctx, fields)
}

// AddToken inserts a token in the database.
func (tokensDB *tokensDB) AddToken(ctx context.Context, token *tokens.UserToken) error {
	query := `INSERT INTO users_tokens (id, user_id, token, expired_at, created_at)
			  VALUES($1, $2, $3, $4, $5)`

	_, err := tokensDB.pool.Exec(ctx, query, token.ID, token.UserID, token.Token, token.ExpiredAt, token.CreatedAt)

	return ErrTokens.Wrap(err)
}

// DeleteToken removes a token from the database.
func (tokensDB *tokensDB) DeleteToken(ctx context.Context, token string) error {
	query := `DELETE FROM users_tokens
	          WHERE token = $1`

	res, err := tokensDB.pool.Exec(ctx, query, token)
	if err != nil {
		return ErrTokens.Wrap(err)
	}

	rowNum := res.RowsAffected()
	if rowNum == 0 {
		return tokens.ErrNoToken.New("")
	}

	return nil
}

// DeleteTokenByUserId removes a token from the database.
func (tokensDB *tokensDB) DeleteTokenByUserId(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users_tokens
	          WHERE user_id = $1`

	res, err := tokensDB.pool.Exec(ctx, query, id)
	if err != nil {
		return ErrTokens.Wrap(err)
	}

	rowNum := res.RowsAffected()
	if rowNum == 0 {
		return tokens.ErrNoToken.New("")
	}

	return nil
}

// ListActiveSessions gets all sessions by user id form the database.
func (tokensDB *tokensDB) ListActiveSessions(ctx context.Context, userID uuid.UUID) (sessions []tokens.UserToken, err error) {
	query := `SELECT id, user_id, token, expired_at, created_at
			  FROM users_tokens
			  WHERE user_id = $1`

	rows, err := tokensDB.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, ErrTokens.Wrap(err)
	}
	defer rows.Close()

	for rows.Next() {
		var session tokens.UserToken
		err := rows.Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiredAt, &session.CreatedAt)
		if err != nil {
			return nil, ErrTokens.Wrap(err)
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteSessionToken removes a token from the database by session id.
func (tokensDB *tokensDB) DeleteSessionToken(ctx context.Context, userId, sessionId uuid.UUID) error {
	query := `DELETE FROM users_tokens
           WHERE user_id = $1 AND id = $2`

	res, err := tokensDB.pool.Exec(ctx, query, userId, sessionId)
	if err != nil {
		return ErrTokens.Wrap(err)
	}

	rowNum := res.RowsAffected()
	if rowNum == 0 {
		return tokens.ErrNoToken.New("")
	}

	return nil
}
