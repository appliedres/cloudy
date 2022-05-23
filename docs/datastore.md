# Datastore

A Datastore is a simplified interface to store and retrieve data from storage providers (like a database or filesystem). There are 2 types of datastores: Binary and JSON. Neither of these is intented to be the fastest or the most capable. This does provide for a "simple" way to store and retrieve items though without a whole lot of overhead or coding. Also, using the same interfaces means that it is easy to switch things out. Additionally, you can add an "Indexer" that can work with the datastore to index your data. For instance, store data in S3 and index with ElasticSearch. 

## Simple Example

```go
import (
    "context"
    "github.com/appliedres/cloudy"
    "github.com/appliedres/cloudy/datastore"
    pg "github.com/appliedres/cloudy-pg"
)

type Pet struct {
    ID string
    Name string
    Type string
    Description string
    Alive bool
}

var DtPet *datastore.Datatype = &storage.Datatype[*Pet]{
	Name:      "Pet",
	Prefix:    "pet",
	DataStore: pg.NewPostgreSqlJsonDataStore[*Pet]("pets", nil),
}

func DoAlot() {
    ctx := cloudy.StartContext()

    // Create a Pet
    dog := &Pet{
        ID: "pet-1234", 
        Name: "Max", 
        Type: "Dog", 
        Description: "Fawn Great Dane, 11yr old, 185lb", 
        Alive: false,
    }

    // Save the Dog
    _, err := DtPet.Save(ctx, dog)
    if err != nil {
        cloudy.Error(ctx, "Could not save dog %v, because %v\n", dog.Name, err )
        return 
    }
    fmt.Printf("Saved %v\n", dog.ID)

    // Load the dog by ID
    dog2, err := DtPet.Get(ctx, "pet-1234")
    if err != nil {
        cloudy.Error(ctx, "Could not load dog %v, because %v\n", "pet-1234", err )
        return 
    }

    // Find the dog using a simple query
    query := datastore.NewQuery()
    query.Conditions.Equals("ID", "pet-1234")

    dogs, err := DtPet.Query(ctx, query)
    if err != nil {
        cloudy.Error(ctx, "Could not query for dogs, because %v\n", err )
        return 
    }

    // Delete the dog
    err = DtPet.Delete(ctx, "pet-1234")
    if err != nil {
        cloudy.Error(ctx, "Could not delete dog %v, because %v\n", "pet-1234", err )
        return 
    }

}
```

The above example shows the a basic, but complete, usage of the data store for pets. It uses the optional idea of data types. In this example we define the data type to have a name, prefix and Data store. The datastore is provided by the PostgreSQL driver. As you can see we use Go Generics where they fit. This makes most operations typesafe.

## Data Type 
The `DataType` object enhances the use of a datastore

## Interceptors
Interceptors provide a way to call code before and after most datatype operations. This can be used as hooks to modify the data (either on gets or saves) or to provide other capablities (such as metrics capture). A good use for this is adding mandatory data that can be calculated. For instance, generating an ID, timestamping, adding the saving user, etc.

## Simple Query
The simple query interface is a basic set of query capablites that should suffice for item and collection level queries. Each driver should provide a translation mechanism from a simple query to a native query. Note that not all drivers have a query implementation and the `BinaryDataStore` does not support queries. A simple query has the following elements: 

```go
type SimpleQuery struct {
	Size       int
	Offset     int
	Colums     []string
	Conditions *SimpleQueryConditionGroup
	SortBy     []*SortBy
}
```
A simple example is

```go
query := datastore.NewQuery()
query.Conditions.Equals("ID", "pet-1234")
```

But it can get more complex

```go
query := datastore.NewQuery()
query.Size = 100
query.Columns = []string{"ID", "Name", "Alive"}

query.Conditions.Includes("ID", []string{"pet-1234", "pet-12345"})

orGroup := query.Conditions.Or()
orGroup.Equals("Alive", "true")
orGroup.Equals("Type", "Dog")
```

## Native Query
If for any reason the simple query does not meet your needs you can always use the native query mechanism. But if you use the native queries then switching between drivers will mean additional code. Here is a basic example of a native query from 
the elastic search driver

```go
es := elastic.NewQuery()
es.Size = 1000
es.Query.Bool.Must.Match("ID.keyword", "pet-1234")

```

Since this is a native query you can take advantage of all the cool technology specific features. In the case of ElasticSearch
this means total hit counting, aggregations, etc. For SQL Databases this would mean joins and groupings

## Datastore Initialization
There are 2 stages of initialization available in the datastore. 

### Table / Index / Storage creation
Some datastores require that the destination for a data store is created and available prior to usage. This stage allows the datastore to create the neccessary items. Listed below are a few details per driver

|Driver|Behavior|
|------|--------|
|Filesystem|Create the necessary directory|
|In Memory|Create the map to store the data|
|PostgreSQL|Create the table and constraints|
|ElasticSearch|Create the Index with any settings (like mappings)|
|Azure - CosmosDB|Create Table|
|Azure - Blob Storage|Create Blob Container|
|AWS - S3|Create Bucket|



