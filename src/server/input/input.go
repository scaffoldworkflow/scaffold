package input

import (
	"fmt"
	"scaffold/server/constants"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scaffold/server/mongodb"
)

type Input struct {
	Name        string `json:"name" bson:"name" yaml:"name"`
	Cascade     string `json:"cascade" bson:"cascade" yaml:"cascade"`
	Description string `json:"description" bson:"description" yaml:"description"`
	Default     string `json:"default" bson:"default" yaml:"default"`
	Type        string `json:"type" bson:"type" yaml:"type"`
}

func CreateInput(i *Input) error {
	ii, err := GetInputByNames(i.Cascade, i.Name)
	if err != nil {
		return fmt.Errorf("error getting inputs: %s", err.Error())
	}
	if ii != nil {
		return fmt.Errorf("input already exists with names %s, %s", i.Cascade, i.Name)
	}

	_, err = mongodb.Collections[constants.MONGODB_INPUT_COLLECTION_NAME].InsertOne(mongodb.Ctx, i)
	return err
}

func DeleteInputByNames(cascade, name string) error {
	filter := bson.M{"cascade": cascade, "name": name}

	collection := mongodb.Collections[constants.MONGODB_INPUT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount != 1 {
		return fmt.Errorf("no input found with names %s, %s", cascade, name)
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

func GetInputByNames(cascade, name string) (*Input, error) {
	filter := bson.M{"cascade": cascade, "name": name}

	inputs, err := FilterInputs(filter)

	if err != nil {
		return nil, err
	}

	if len(inputs) == 0 {
		return nil, nil
	}

	if len(inputs) > 1 {
		return nil, fmt.Errorf("multiple inputs found with names %s, %s", cascade, name)
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

func UpdateInputByNames(cascade, name string, i *Input) error {
	filter := bson.M{"cascade": cascade, "name": name}

	collection := mongodb.Collections[constants.MONGODB_INPUT_COLLECTION_NAME]
	ctx := mongodb.Ctx

	opts := options.Replace().SetUpsert(true)

	result, err := collection.ReplaceOne(ctx, filter, i, opts)

	if err != nil {
		return err
	}

	if result.ModifiedCount != 1 {
		return fmt.Errorf("no input found with names %s, %s", cascade, name)
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

	return inputs, nil
}
