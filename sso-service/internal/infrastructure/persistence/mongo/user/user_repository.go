package usermongo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/devathh/coderun/sso-service/internal/domain/user"
	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoRepository struct {
	log        *slog.Logger
	client     *mongo.Client
	collection *mongo.Collection
}

func New(log *slog.Logger, client *mongo.Client) (*MongoRepository, error) {
	if log == nil || client == nil {
		return nil, customerrors.ErrNilArgs
	}

	return &MongoRepository{
		log:        log,
		client:     client,
		collection: client.Database("shost").Collection("users"),
	}, nil
}

func (mr *MongoRepository) Save(ctx context.Context, user *user.User) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if user == nil {
		return nil, customerrors.ErrNilArgs
	}

	if _, err := mr.collection.InsertOne(ctx, toModel(user)); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, customerrors.ErrUserAlreadyRegistered
		}

		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (mr *MongoRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return nil, customerrors.ErrNilArgs
	}

	var model UserModel
	if err := mr.collection.FindOne(ctx, bson.M{"email": email}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, customerrors.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user, err := toDomain(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return user, nil
}

func (mr *MongoRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var model UserModel
	if err := mr.collection.FindOne(ctx, bson.M{"id": id}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, customerrors.ErrUserDoesntExist
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	user, err := toDomain(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return user, nil
}

func (mr *MongoRepository) Update(ctx context.Context, updUser *user.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	filter := bson.M{"id": updUser.ID().String()}
	update := bson.M{"$set": bson.M{"username": updUser.Username()}}

	res, err := mr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if res.MatchedCount == 0 {
		return customerrors.ErrUserDoesntExist
	}

	return nil
}
