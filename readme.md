# Cloudy
![badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/wtiger001/3652cbbb7e810afda7a001bf1859e16c/raw/cloudy-coverage.json)

IN WORK -- DO NOT USE 

Planned to be an open source library that is a runtime adapter for AWS and Azure cloud capablilies. This library tries to provide simple capalbities instead of trying for a full-coverage. As you ned to use more complex capalbities they you should be able to tie into the underlying APIs from each of the Cloud Vendors.

## Capablities

- User Management
- Group Management
- Blob / Bucket 
- Mail
- SMS / Notification
- JSON Data storage / query
- Binary Data storage

## Providers

### AWS
Amazon Web Services 
- User Management - Cognito
- Group Management - Cognito
- Blob / Bucket - S3
- Mail - SES
- SMS - SNS
- JsonDataStorage - OpenSearch
- BinaryDataStorage - S3

### Azure
Microsoft Azure
- User Management - AAD
- Group Management - AAD
- Blob / Bucket - Azure Blob Storage
- Mail - Azure Communication Services
- SMS - Azure Communication Services
- JsonDataStorage - Cosmos
- BinaryDataStorage - Azure Blob Storage

### ElasticSearch
- JsonDataStorage

### Keycloak
- User Management
- Group Management
