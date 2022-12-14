package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

// custom mongodb oauth client store for go-oauth2
// non exported client model as it is only required for internal implementation
// implements oauth2.ClientStore with additional operations for CRUD

const (
	key_client_id = "user_id"
)

type OAuthClientStorer interface {
	oauth2.ClientStore
	Set(info oauth2.ClientInfo) error
	RemoveByID(id string) error
}

type MongoClientStore struct {
	dbclient   *mongo.Client
	database   string
	collection string
}

type client struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Secret string             `bson:"secret"`
	Domain string             `bson:"domain"`
	UserID string             `bson:"user_id"`
}

func NewMongoClientStore(dbclient *mongo.Client, dbname string) *MongoClientStore {
	return &MongoClientStore{dbclient: dbclient, database: dbname, collection: "oauth_client"}
}

// Save a client
func (cs *MongoClientStore) Set(info oauth2.ClientInfo) error {
	coll := cs.dbclient.Database(cs.database).Collection(cs.collection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data := client{
		// for now we will only allow generatedIds...
		Secret: info.GetSecret(),
		Domain: info.GetDomain(),
		UserID: info.GetUserID(),
	}
	_, err := coll.InsertOne(ctx, data)
	return err
}

// GetByID according to the ID for the client information
func (cs *MongoClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	var cd client
	var res models.Client
	coll := cs.dbclient.Database(cs.database).Collection(cs.collection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := coll.FindOne(ctx, bson.M{key_client_id: id}).Decode(&cd)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrorNoResult
		}
		return nil, err
	}
	// doesn't seem right to expose this ID in the domain model
	res.ID = cd.ID.String()
	res.UserID = cd.UserID
	res.Domain = cd.Domain
	res.Secret = cd.Secret
	return &res, nil
}

// Delete a client by the id
func (ts *MongoClientStore) RemoveByID(id string) error {
	coll := ts.dbclient.Database(ts.database).Collection(ts.collection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := coll.DeleteOne(ctx, bson.M{key_client_id: id})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
	}
	return err
}
