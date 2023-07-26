package input

import (
	"fmt"
	"scaffold/server/constants"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type Input struct {
	Name        string `json:"name" bson:"name"`
	Cascade     string `json:"cascade" bson:"cascade"`
	Description string `json:"description" bson:"description"`
	Default     string `json:"default" bson:"default"`
	Type        string `json:"type" bson:"type"`
}

func CreateInput(i *Input) error {
	if _, err := GetInputByNames(i.Cascade, i.Name); err == nil {
		return fmt.Errorf("input already exists with names %s, %s", i.Cascade, i.Name)
	}

	_, err := mongodb.Collections[constants.MONGODB_INPUT_COLLECTION_NAME].InsertOne(mongodb.Ctx, i)
	return err
}

func DeleteInputByNames(cascade, input string) error {
	filter := bson.M{"cascade": cascade, "input": input}

	collection := mongodb.Collections[constants.MONGODB_INPUT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no input found with names %s, %s", cascade, input)
	}

	return nil

}

func DeleteInputsByCascade(cascade string) error {
	filter := bson.M{"cascade": cascade}

	collection := mongodb.Collections[constants.MONGODB_INPUT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteMany(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no inputs found with cascade %s", cascade)
	}

	return nil

}

func GetAllInputs() ([]*Input, error) {
	filter := bson.D{{}}

	inputs, err := FilterInputs(filter)

	return inputs, err
}

func GetInputByNames(cascade, input string) (*Input, error) {
	filter := bson.M{"cascade": cascade, "input": input}

	inputs, err := FilterInputs(filter)

	if err != nil {
		return nil, err
	}

	if len(inputs) == 0 {
		return nil, fmt.Errorf("no input found with names %s, %s", cascade, input)
	}

	if len(inputs) > 1 {
		return nil, fmt.Errorf("multiple inputs found with names %s, %s", cascade, input)
	}

	return inputs[0], nil
}

func GetInputsByCascade(cascade string) ([]*Input, error) {
	filter := bson.M{"cascade": cascade}

	inputs, err := FilterInputs(filter)

	if err != nil {
		return nil, err
	}

	return inputs, nil
}

func UpdateInputByNames(cascade, input string, i *Input) error {
	filter := bson.M{"cascade": cascade, "input": input}

	collection := mongodb.Collections[constants.MONGODB_INPUT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, i, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no input found with names %s, %s", cascade, input)
	}

	return nil
}

func FilterInputs(filter interface{}) ([]*Input, error) {
	// A slice of inputs for storing the decoded documents
	var inputs []*Input

	collection := mongodb.Collections[constants.MONGODB_INPUT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return inputs, err
	}

	for cur.Next(ctx) {
		var s Input
		err := cur.Decode(&s)
		if err != nil {
			return inputs, err
		}

		inputs = append(inputs, &s)
	}

	if err := cur.Err(); err != nil {
		return inputs, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(inputs) == 0 {
		return inputs, mongo.ErrNoDocuments
	}

	return inputs, nil
}
