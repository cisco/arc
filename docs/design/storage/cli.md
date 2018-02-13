# Storage CLI Design

Three new commands will be added to the amp cli to provide management of storage.


## Storage

To run an amp command against all buckets and bucket sets in an account use the form

```shell
amp [account_name] storage [command] [flags]
```

The commands available to storage are

- **update**: Update the tags for the entirety of the account storage.
- **audit**: Audits the entirety of the account storage.
- **info**: Provides run time information for the storage.
- **config**: Provides configuration information for the storage.
- **help**: Provides help with the storage command.


For example, to run an audit command against the accoount in the fictional "example-integration" account, run the command

```shell
amp example-integration storage audit
```


## Named bucket

To run a command against a single bucket in storage use the form

```shell
amp [account_name] bucket [bucket_name] [command] [flags]
```

where "bucket_name" is the name of the bucket.


The commands available to a bucket are

- **create**: Creates the bucket. If the bucket already exists this command will do nothing.
- **destroy**: Destroys the bucket. If the bucket does not exist this command will do nothing.
- **update**: Updates the tags for this bucket. If the bucket does not exist this command will do nothing.
- **audit**: Audit the bucket. The audit will compare the configuration of the and report if the run time buckets do not match.
- **info**: Provides run time information for the bucket.
- **config**: Provides configuration information for the bucket.
- **help**: Provides help with the named bucket command.


For example, to create a bucket named "contacts" in the "example-integration" account, run the command

```shell
amp example-integration bucket contacts create
```


## Named bucket set

To run a command against a single bucket set in storage use the form

```shell
amp [account_name] bucket_set [bucket_set_name] [command] [flags]
```

where "bucket_set_name" is the name of the bucket set.

The commands available to a bucket set are

- **create**: Creates the bucket set and enables replication for each bucket in the bucket set. If the bucket set already exists this command will do nothing.
- **destroy**: Destroys the bucket set. If the bucket set does not exist the command will do nothing.
- **update**: Updates the tags on each bucket in the bucket set. If the bucket set does not exist this command will do nothing.
- **audit**: Audits each bucket within the bucket set. The audit will compare the configuration of the bucket set and report if there are run time buckets that do not match.
- **info**: Provides run time information for the bucket set.
- **config**: Provides configuration information for the bucket set.
- **help**: Provides help with the named bucket set command.
