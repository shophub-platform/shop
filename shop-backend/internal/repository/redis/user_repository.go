package redis

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

// Key schema:
//
//	user:{id}          — HASH  (id, email, password_hash, role, created_at)
//	user:email:{email} — STRING → id  (index for lookup by email)

func userKey(id uuid.UUID) string    { return "user:" + id.String() }
func userEmailKey(email string) string { return "user:email:" + email }

type userRepository struct {
	rdb *redis.Client
}

func NewUserRepository(rdb *redis.Client) repository.UserRepository {
	return &userRepository{rdb: rdb}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	emailKey := userEmailKey(user.Email)

	// Check duplicate email atomically
	exists, err := r.rdb.Exists(ctx, emailKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return repository.ErrAlreadyExists
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, userKey(user.ID),
		"id", user.ID.String(),
		"email", user.Email,
		"password_hash", user.PasswordHash,
		"role", user.Role,
		"created_at", now.Format(time.RFC3339),
		"updated_at", now.Format(time.RFC3339),
	)
	pipe.Set(ctx, emailKey, user.ID.String(), 0)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	idStr, err := r.rdb.Get(ctx, userEmailKey(email)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, repository.ErrNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	fields, err := r.rdb.HGetAll(ctx, userKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, repository.ErrNotFound
	}

	uid, _ := uuid.Parse(fields["id"])
	createdAt, _ := time.Parse(time.RFC3339, fields["created_at"])
	updatedAt, _ := time.Parse(time.RFC3339, fields["updated_at"])

	return &model.User{
		ID:           uid,
		Email:        fields["email"],
		PasswordHash: fields["password_hash"],
		Role:         fields["role"],
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}
