package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/Av1shay/di-demo/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const itemCollName = "items"

type Repository struct {
	client *mongo.Client
	dbName string
}

func NewRepository(uri, dbName string) (*Repository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &Repository{
		client: client,
		dbName: dbName,
	}, nil
}

func (r *Repository) GetItemByName(ctx context.Context, name string) (types.Item, error) {
	collection := r.client.Database(r.dbName).Collection(itemCollName)

	var item Item
	err := collection.FindOne(ctx, bson.M{"name": name}).Decode(&item)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return types.Item{}, &types.APIError{
				Code: types.ErrorCodeNotFound,
				Msg:  fmt.Sprintf("item with name %s not found", name),
			}
		}
		return types.Item{}, err
	}

	return parseItem(item), nil
}

func (r *Repository) SaveItem(ctx context.Context, input types.ItemCreateInput) (types.Item, error) {
	collection := r.client.Database(r.dbName).Collection(itemCollName)

	item := Item{
		ID:    primitive.NewObjectID(),
		Name:  input.Name,
		Value: input.Value,
	}
	_, err := collection.InsertOne(ctx, item)
	if err != nil {
		return types.Item{}, err
	}

	return r.GetItemByName(ctx, input.Name)
}

func (r *Repository) ListItems(ctx context.Context) ([]types.Item, error) {
	collection := r.client.Database(r.dbName).Collection(itemCollName)

	cur, err := collection.Find(ctx, bson.D{})
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

func (r *Repository) DeleteItem(ctx context.Context, id string) error {
	_, err := r.client.Database(r.dbName).Collection(itemCollName).DeleteOne(ctx, bson.D{{"_id", id}})
	return err
}

func (r *Repository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

func parseItem(item Item) types.Item {
	return types.Item{
		ID:        item.ID.Hex(),
		Name:      item.Name,
		Value:     item.Value,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
