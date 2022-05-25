# Cloudy Documentation

Cloudy is a go-lang library that is used to help alleviate the need to write cloud vendor specific code. It makes the basic assumption that in most cases you only need a relatively simple interaction with the cloud services and you do not need to utilize every bell and whistle offered. To this end, Cloudy has some basic interaces and helper code to interface with a variety of providers. It takes some inspiration from the database/sql package in the go libraries. 

## Explicit vs Dynamic
In order to get an appropriate driver for a cloudy managed provider there are typically two ways. The first way is the explicit way. In this way we directly use the concrete instantiations of the Cloudly Providers. This allows us to easily, and in a type-safe manner, instantiate and call the provider. This way also lets us easily access capablities for the cloud service that are not exposed in the simpler, cloud-nuetral provider interface. The downside of this approach is that it can take a bit of refactoring to change out providers and it difficult to make the provider change at runtime (during an initialization phase for instance)

### Explicit Usage
An example of explicit usage is shown below with an ElasticSearch Datastore
```go

import (
   	"github.com/appliedres/cloudy"
   	"github.com/appliedres/cloudy-elastic"
)

// Localhost connection to ElasticSearch without a username or password
var info = &cloudyelastic.ConnectionInfo{
	Endpoint: "http://localhost:9201",
}

ctx := cloudy.StartContext()
ds := cloudyelastic.NewElasticJsonDataStore[tests.TestItem](
    "test",
)

err := ds.Open(ctx, info)
if err != nil {
    panic(err)
}

ds.Client.CreateIndex()

```

In this above case we have to explicitly import the `cloudyelastic` packaage and we directly use the configuration object `ConnectionInfo` but we are still using the generic `JsonDataStore` interface. If we need to perform action specific to the ElasticSearch API we can retrive the `Client` object from the `ds` instance. But if we do that then obviously we are now unable to easly switch out the datastore for a different implementation. 

### Dynamic Usage
An alternative approach to the explict use is the dynamic or implicit one. This approach should be used when it is important to dynamically find the correct driver an use that. For instance, if you have a system that can be configured at runtime to select the correct backend for a user manager, datastore, etc. Then using the dynamic approach will just require some configuration without code changes. This approach was patterned after the way datbase/sql was designed. An example of this is shown below. 

```go

import (
    "fmt"
    
    "github.com/appliedres/cloudy"
    
    // NOTE we have to still reference the aws package some where in our code so that the `init()` function
    // calls. But we never explicitly use the package so we have to put the `_` in front of it.
    _ "github.com/appliedres/cloudy-aws" 
)

// Load some configuration from the environment
providerName, configurationMap, err := loadFromEnv(ctx)
if err != nil {
    panic(err)
}

ctx := cloudy.StartContext()
mgr, err := cloudy.UserProviders.New(providerName string, configurationMap)
if err != nil {
    panic(err)
}

user, err := mgr.GetUser(ctx, "myuserID")
if err != nil {
    panic(err)
}

fmt.Printf("Found %+v\n", user)
```

In the above example we show the dynamic instantiation. We load the configuration information from the environment variables (not shown) and the output is the name of the provider and a map of the configuration. The required keys / values for the map vary per provider and are documented in each provider. Then we can access the `UserProviders` service that is in the `cloudy` package. This service is a registry of all providers that have self registered with an `init()`function in their package. The provider is then created with the factory method that is supplied by the provider and the generic interface is returned. 


