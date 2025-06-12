/*
Packages storage manages storage in cloud services. This is focused primarily on the
storage of blobs / objects.

# Description
There are three primary concepts: <Account>, <Container>, <Item>. The <Account> interfaces
are used to represent a type of storage account. A storage account can container multiple
<Container> items. A container item represents something that can hold items. Items are
the things / objects that are stored.

A few examples are:

# Azure Blob Storage
StorageAccount -> BlobContainers -> BlobItems or PrefixItems

# AWS S3
S3Account -> Bucket -> Object

# Sample Usage

	// Load the account
	account, err := storage.ObjectStorageAccount.NewFromEnv(env.Segment("PERSONAL_FILE_SHARE"), "DRIVER")
	if err != nil {
		log.Fatalf("Failed to create storage account: %v\n", err)
	}

	fmt.Printf("Account Name: ", account.Name())

	// Get the containers
	containers, err := account.List(ctx, &storage.ListContainerOptions{
		pageToken: nil,
	})

	if err != nil {
		log.Fatalf("Failed to list the containers in an account: %v\n", err)
	}

	for _, container := range containers {

		items, prefixes, err := container.List(ctx, &storage.ListItemsOptions{
			pageToken: nil,
			prefix: "",
		})

		for _, p := range prefixes {
			fmt.Printf("%v\n", p.Name())
		}

		for _, item := range items {
			fmt.Printf("%v\n", item.Name())
			for k, v := range item.Tags() {
				fmt.Printf("\t%v:%v\n", k, v)
			}
		}
	}

# Current Implementations

- Filesystem
  - Directory (rw) - Local filesystem storage
  - Zip (rw) - ZIP archive storage
  - Tar (rw) - TAR/GZIP archive storage

# Planned Implementations

- Docker Registry (ro) - Access files in container images directly
- Artifactory (rw) - JFrog Artifactory storage
- PGSql (rw) - PostgreSQL database storage
- HTTP/URL Storage (ro) - Simple downloads from HTTP URLs

# Other Implementations
- AWS S3 (rw) - Amazon Simple Storage Service and compatible (MinIO, Wasabi, etc.)
- Azure Blob Storage (rw) - Microsoft's blob storage offering
- Git Repository (ro) - Access files from Git repositories

## Potential Additional Implementations

## Cloud Providers
- Google Cloud Storage (rw) - Google's object storage service
- Oracle Cloud Storage (rw) - Oracle's object storage service

## Protocol-based Storage
- SFTP/FTP Storage (rw) - Remote file access via SFTP protocol
- WebDAV Storage (rw) - HTTP-based file manipulation

## Database Storage
- SQL Database Storage (rw) - Store blobs in SQL databases
- NoSQL Storage (rw) - MongoDB, DynamoDB storage backend
- Key-Value Stores (rw) - Redis, etcd based storage

## Specialized Storage
- Content Addressable Storage (rw) - IPFS, Arweave
- Email Attachment Storage (ro) - Extract from email accounts
- Cache Storage (rw) - Temporary high-speed storage
- In-Memory Storage (rw) - Volatile RAM-based storage

## Meta Implementations
- Union Storage (rw) - Combine multiple backends
- Encrypted Storage (rw) - Add encryption to any storage backend
- Mirrored Storage (rw) - Replicate across multiple backends
- Sharded Storage (rw) - Split data across multiple backends
- Versioned Storage (rw) - Add version history to any backend

# Implementation Considerations

When implementing new storage backends, consider:
- Authentication mechanisms
- Rate limiting and throttling
- Error handling and retries
- Performance characteristics
- Atomic operation support
- Metadata capabilities
- Content addressing vs. location addressing
*/
package storage
