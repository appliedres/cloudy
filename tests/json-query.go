package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/appliedres/cloudy/datastore"
	"github.com/stretchr/testify/assert"
)

type TestQueryItem struct {
	ID     string
	Name   string
	Val1   int
	Val2   int
	Date1  time.Time
	Date2  time.Time
	StrArr []string
	ObjArr []*TestQueryItemChild
}

type TestQueryItemChild struct {
	Name string
}

func QueryJsonDataStoreTest(t *testing.T, ctx context.Context, ds datastore.JsonDataStore[TestQueryItem]) {
	// Create the item
	testItem := &TestQueryItem{
		ID:     "TEST-12345",
		Name:   "My Test Item",
		Val1:   23,
		Val2:   45,
		Date1:  time.Now(),
		Date2:  time.Now().Add(3 * time.Hour),
		StrArr: []string{"One", "Two", "Three"},
		ObjArr: []*TestQueryItemChild{
			{
				Name: "Obj1",
			},
			{
				Name: "Obj2",
			},
		},
	}

	err := ds.Save(ctx, testItem, testItem.ID)
	assert.Nil(t, err, "Should not get an error saving to the database")

	qById := datastore.NewQuery()
	qById.Conditions.Equals("ID", "TEST-12345")

	results, err := ds.Query(ctx, qById)
	assert.Nil(t, err, "ID - Should not get an error saving to the database")
	assert.Equal(t, len(results), 1, "ID - Should get one result returned")

	qByName := datastore.NewQuery()
	qByName.Conditions.Equals("Name", "My Test Item")

	results, err = ds.Query(ctx, qByName)
	assert.Nil(t, err, "Name - Should not get an error saving to the database")
	assert.Equal(t, len(results), 1, "Name - Should get one result returned")

	qByVal1 := datastore.NewQuery()
	qByVal1.Conditions.Equals("Val1", "23")

	results, err = ds.Query(ctx, qByVal1)
	assert.Nil(t, err, "Val Equal - Should not get an error saving to the database")
	assert.Equal(t, len(results), 1, "Val Equal - Should get one result returned")

	qRange := datastore.NewQuery()
	qRange.Conditions.Between("Val1", "0", "100")

	results, err = ds.Query(ctx, qRange)
	assert.Nil(t, err, "Between - Should not get an error during query")
	assert.Equal(t, len(results), 1, "Between - Should get one result returned")

	qContains := datastore.NewQuery()
	qContains.Conditions.Contains("StrArr", "One")

	results, err = ds.Query(ctx, qContains)
	assert.Nil(t, err, "Contains - Should not get an error during query")
	assert.Equal(t, len(results), 1, "Contains - Should get one result returned")

	qAnd := datastore.NewQuery()
	qAnd.Conditions.Contains("StrArr", "One")
	qAnd.Conditions.Equals("ID", "TEST-12345")

	results, err = ds.Query(ctx, qAnd)
	assert.Nil(t, err, "And - Should not get an error during query")
	assert.Equal(t, len(results), 1, "And - Should get one result returned")

	qLt := datastore.NewQuery()
	qLt.Conditions.LessThan("Val1", "100")

	results, err = ds.Query(ctx, qLt)
	assert.Nil(t, err, "Less Than - Should not get an error during query")
	assert.Equal(t, len(results), 1, "Less Than - Should get one result returned")

	qComposite := datastore.NewQuery()
	qComposite.Conditions.Contains("StrArr", "One")
	qComposite.Conditions.Equals("ID", "TEST-12345")
	gGrpOr := qComposite.Conditions.Or()
	gGrpOr.Equals("Val1", "23")
	gGrpOr.Equals("Val1", "24")

	results, err = ds.Query(ctx, qComposite)
	assert.Nil(t, err, "Composite - Should not get an error during query")
	assert.Equal(t, len(results), 1, "Composite - Should get one result returned")

	fmt.Println("Done")
}
