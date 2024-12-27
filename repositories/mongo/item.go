package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/Av1shay/di-demo/pkg/errs"
	"github.com/Av1shay/di-demo/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

const itemCollName = "items"

var orderBys = map[types.OrderBy]string{
	types.OrderByName:      "name",
	types.OrderByCreatedAt: "created_at",
	types.OrderByUpdatedAt: "updated_at",
}

type Item struct {
	ID        primitive.ObjectID `bson:"_id"`
	AccountID string             `bson:"account_id"`
	Name      string             `bson:"name"`
	Value     string             `bson:"value"`
	Version   int                `bson:"version"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (r *Repository) GetItemByName(ctx context.Context, name string, accountID string) (types.Item, error) {
	var item Item
	err := r.itemsColl.FindOne(ctx, bson.M{"name": name, "account_id": accountID}).Decode(&item)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return types.Item{}, &errs.AppError{
				Code: errs.ErrorCodeNotFound,
				Msg:  fmt.Sprintf("item '%s' not found", name),
				Err:  err,
			}
		}
		return types.Item{}, err
	}

	return parseItem(item), nil
}

func (r *Repository) SaveItem(ctx context.Context, input types.ItemCreateInput, accountID string) (types.Item, error) {
	now := time.Now()
	item := Item{
		ID:        primitive.NewObjectID(),
		AccountID: accountID,
		Name:      input.Name,
		Value:     input.Value,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := r.itemsColl.InsertOne(ctx, item)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return types.Item{}, &errs.AppError{
				Code: errs.ErrorCodeDuplicate,
				Msg:  fmt.Sprintf("item '%s' already exist", input.Name),
				Err:  err,
			}
		}
		return types.Item{}, err
	}

	return r.GetItemByName(ctx, input.Name, accountID)
}

func (r *Repository) UpdateItem(ctx context.Context, input types.UpdateItemInput, accountID string) (types.Item, error) {
	objID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		return types.Item{}, fmt.Errorf("failed to convert item id to ObjectID: %w", err)
	}
	filter := bson.D{{"_id", objID}, {"account_id", accountID}}
	update := bson.M{
		"$set": bson.M{
			"name":       input.Name,
			"value":      input.Value,
			"updated_at": time.Now(),
		},
		"$inc": bson.M{
			"version": 1,
		},
	}
	_, err = r.itemsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return types.Item{}, err
	}
	return r.GetItemByName(ctx, input.Name, accountID)
}

func (r *Repository) ListItems(ctx context.Context, input types.ListItemsInput, accountID string) ([]types.Item, error) {
	opts := options.Find()

	sortOrder := 1
	if input.Sort == types.DESC {
		sortOrder = -1
	}
	orderBy := "updated_at"
	if ob, ok := orderBys[input.OrderBy]; ok {
		orderBy = ob
	}

	opts.SetSort(bson.D{{orderBy, sortOrder}})

	if input.Limit > 0 {
		opts.SetLimit(int64(input.Limit))
	}

	cur, err := r.itemsColl.Find(ctx, bson.D{{"account_id", accountID}}, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	defer cur.Close(ctx)

	var res []types.Item
	for cur.Next(ctx) {
		var item Item
		err := cur.Decode(&item)
		if err != nil {
			return nil, err
		}
		res = append(res, parseItem(item))
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) DeleteItem(ctx context.Context, id string, accountID string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("failed to convert item id to ObjectID: %w", err)
	}
	_, err = r.itemsColl.DeleteOne(ctx, bson.D{
		{"_id", objID},
		{"account_id", accountID},
	})
	return err
}

func parseItem(item Item) types.Item {
	return types.Item{
		ID:        item.ID.Hex(),
		Name:      item.Name,
		Value:     item.Value,
		Version:   item.Version,
		AccountID: item.AccountID,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
