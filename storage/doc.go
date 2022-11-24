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
*/
package storage
