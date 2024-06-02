package webhook

import (
	"fmt"
	"scaffold/server/constants"
	"scaffold/server/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type Webhook struct {
	ID         string `json:"id" bson:"id" yaml:"id"`
	Entrypoint string `json:"entrypoint" bson:"entrypoint" yaml:"entrypoint"`
	Cascade    string `json:"cascade" bson:"cascade" yaml:"cascade"`
}

func CreateWebhook(w *Webhook) error {
	id := utils.GenerateToken(64)

	w.ID = id

	_, err := mongodb.Collections[constants.MONGODB_WEBHOOK_COLLECTION_NAME].InsertOne(mongodb.Ctx, w)
	return err
}

func DeleteWebhookByID(id string) error {
	filter := bson.M{"id": id}

	collection := mongodb.Collections[constants.MONGODB_WEBHOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no webhook found with id %s", id)
	}

	return nil

}

func GetAllWebhooks() ([]*Webhook, error) {
	filter := bson.D{{}}

	webhooks, err := FilterWebhooks(filter)

	return webhooks, err
}

func GetWebhookByID(id string) (*Webhook, error) {
	filter := bson.M{"id": id}

	webhooks, err := FilterWebhooks(filter)

	if err != nil {
		return nil, err
	}

	if len(webhooks) == 0 {
		return nil, nil
	}

	if len(webhooks) > 1 {
		return nil, fmt.Errorf("multiple webhooks found with id %s", id)
	}

	return webhooks[0], nil
}

func UpdateWebhooksByID(id string, w *Webhook) error {
	filter := bson.M{"id": id}

	collection := mongodb.Collections[constants.MONGODB_WEBHOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, w, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no webhook found with id %s", id)
	}

	return nil
}

func FilterWebhooks(filter interface{}) ([]*Webhook, error) {
	// A slice of tasks for storing the decoded documents
	var webhooks []*Webhook

	collection := mongodb.Collections[constants.MONGODB_WEBHOOK_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return webhooks, err
	}

	for cur.Next(ctx) {
		var w Webhook
		err := cur.Decode(&w)
		if err != nil {
			return webhooks, err
		}

		webhooks = append(webhooks, &w)
	}

	if err := cur.Err(); err != nil {
		return webhooks, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	return webhooks, nil
}
