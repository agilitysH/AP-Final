package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsureIndexes(ctx context.Context) error {
	if err := ensureUsersIndexes(ctx); err != nil {
		return err
	}
	if err := ensureEnrollmentsIndexes(ctx); err != nil {
		return err
	}
	if err := ensureProgressIndexes(ctx); err != nil {
		return err
	}
	return nil
}

func ensureUsersIndexes(ctx context.Context) error {
	_, err := GetCollection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func ensureEnrollmentsIndexes(ctx context.Context) error {
	_, err := GetCollection("enrollments").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "courseId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "courseId", Value: 1}, {Key: "status", Value: 1}},
		},
	})
	return err
}

func ensureProgressIndexes(ctx context.Context) error {
	_, err := GetCollection("progress").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "courseId", Value: 1}, {Key: "itemId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}
