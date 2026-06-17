package store

import (
	"context"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestSeedAddsNewDomainPermissionsToExistingAdmins(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("updates existing admin permissions before returning", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(1)}}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}),
		)

		if err := store.Seed(context.Background()); err != nil {
			mt.Fatalf("Seed error: %v", err)
		}
		_ = mt.GetStartedEvent()
		update := mt.GetStartedEvent()
		if update == nil || update.CommandName != "update" {
			mt.Fatalf("expected admin permission update, got %#v", update)
		}
		command := update.Command.String()
		for _, permission := range []string{"review_notes:write", "policy:write"} {
			if !strings.Contains(command, permission) {
				mt.Fatalf("update command missing %s: %s", permission, command)
			}
		}
	})
}
