# Configuration

I am not really happy with this solution yet. I am looking for a good way to heirachically load configuration from either environment variables, secrets, files or another service.

### With Envrionment variables
```bash
CLOUDY_PREFIX="PREFIX"
PREFIX_USERDOMAIN="somwhere.microsoft.us"

PREFIX_USERS_DRIVER="azure-msgraph"
PREFIX_USERS_TENANT_ID="asdasda"
PREFIX_USERS_CLIENT_ID="asdasda"
PREFIX_USERS_CLIENT_SECRET="asdasda"

PREFIX_GROUP_DRIVER="azure-msgraph"
PREFIX_GROUP_TENANT_ID="asdasda"
PREFIX_GROUP_CLIENT_ID="asdasda"
PREFIX_GROUP_CLIENT_SECRET="asdasda"

PREFIX_PRIMARY_VMS_DRIVER="azure"
PREFIX_PRIMARY_VMS_TENANT_ID="foo"
PREFIX_PRIMARY_VMS_CLIENT_ID="bar"
PREFIX_PRIMARY_VMS_CLIENT_SECRET="different"
```

When the app starts it will look for "CLOUD_PREFIX" Variable or the prefix can be assigned programatically. Then the LoadFromEnv method will take the prefix and the driver name. So for instance:
It would take "PREFIX" and "DRIVER" for the `UserManager` and it would send the following map:

```go
[DRIVER]="azure"
[TENANT_ID]="asdasda"
[CLIENT_ID]="asdasda"
[CLIENT_SECRET]="asdasda"
```
with each block being configured separately. Alternatively, it could be configured as below.

```bash
MSGRAPH_DRIVER="azure-msgraph"
MSGRAPH_AZ_TENANT_ID="asdasda"
MSGRAPH_AZ_CLIENT_ID="asdasda"
MSGRAPH_AZ_CLIENT_SECRET="asdasda"
```

This would configure both user and group to be an MSGraph implementation with the same configuration

### With Files
I would like the same idea as above but with files... e.g. get a part of a file and send that. So for TOML

```toml
[user-manager]
driver=azure-msgraph
tenant-id="asdasda"
client-id="asdasda"
client-secret="asdasda"
```

```json
{
    "user-manager": {
        "driver": "azure-msgraph",
        "tenant-id":"asdasda",
        "client-id":"asdasda",
        "client-secret":"asdasda"
    }
}
```

would produce with `loadFromToml("user-manager", "driver")` or `loadFromJson("user-manager", "driver")`
```go
[DRIVER]="azure"
[TENANT_ID]="asdasda"
[CLIENT_ID]="asdasda"
[CLIENT_SECRET]="asdasda"
```

### Configuration provider
Provides configuration maps...
```go
type ConfigurationProvider interface {
    Load(nameOrPrefix string) (map[string]interface{}, error)
}
```
 A service can then be implemented by the developers. We would have the `Environment`, `Toml`, `Json` services in the core cloudy. Then each datastore or cloud provider could include their own. so for Azure we could have a `KeyVaultConfigProvider`, in AWS we could have a `SecretsManagerConfigProvider`, etc. 

### Questions
- Should we allow default / overrides. For example. Load the config file first but allow the env to override some values?
