package generate

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ObjectId() string {
	return fmt.Sprintf("{\"$oid\":\"%s\"}", primitive.NewObjectID().Hex())
}
