# Cloudy

Planned to be an open source library that is a runtime adapter for AWS and Azure cloud capablilies. This library tries to provide simple capalbities instead of trying for a full-coverage. As you ned to use more complex capalbities they you should be able to tie into the underlying APIs from each of the Cloud Vendors. 

## Structure / Technical Approach

We use interfaces and common models through out this library

## Authentication
Most clouds have similiar authentication strategies. For Azure there is a TenantID, ClientID and Client Secret, for AWS it is a AccessKey and secret

```
client, err := cloudy.NewClient(ctx, tenantId, clientId, clientSecret, nil)

```

## Users

``` 
    users := client.Users()
    user, err := users.NewUser(ctx, userModel)

```

## Groups

``` 
    groups := client.Groups()
    user, err := users.GetGroupsForUser(ctx, userIdorName)

```

## Storage

```
    storage := client.Storage()
    storage.NewAccount()
    storage.NewBucket()

```

## Configuration
- Library wide defaults 
-- Region. API paths, etc

## Models
- Models are documented in OpenAPI V2 spec and compiled with Go-Swagger. These definitions can then be incorporated in other swagger files. 
- How to handle "more fields"? Either: Map or AllOf...
- With additional fields... How to provide the field mapping?
-- Maybe... 
--- Azure.User.Fields['job-title']="JobTitle" and use reflection?  
--- just an adapter.. azure.ToAzModel(user) AzureUzer.. I like #2
--- Only interface (all getters / setters).. ick
--- Additional object added to base struct