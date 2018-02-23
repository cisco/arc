# Key Management CLI Design

Two new commands will be added to the amp cli to provide management of encryption keys.


## Key Management

To run an amp command against all encryption keys in an account use the form

```shell
amp [account_name] key_management [command] [flags]
```

The commands available to key_management are

- **update**: Updates the tags and alias for each key.
- **audit**: Audits the entirety of key management.
- **info**: Provides run time information for the key management.
- **config**: Provides configuration information for the key management.
- **help**: Provides help with the key_management command.


For example, to run an audit command against the key_management in the fictional "example-integration" account, run the command

```shell
amp example-integration key_management audit
```


## Name Encryption Key

To run a command against a single encryption key in key_management use the form

```shell
amp [account_name] encryption_key [key_name] [command] [flags]
or
amp [account_name] key [key_name] [command] [flags]

where "key_name" is the alias to the encryption key.

The commands available to an encryption key are

- **create**: Creates the encryption key. If the encryption key already exists this command will do nothing.
- **destroy**: Schedules the encryption key for deletion. If the encryption key does not exist this command will do nothing.
- **update**: Updates the tags for the encryption key. If the encryption key does not exist this command will do nothing.
- **audit**: Audit the encryption key. This will compare the configuration encryption key and report if the run time encryption keys do not match.
- **info**: Provides run time information for the encryption key.
- **config**: Provides configuration information for the encryption key.
- **help**: Provides help with the named encryption key command.

For example, to create an encryption key with the alias "bucket_replication_key" in the "example-integration" account, run the command

```shell
amp example-integration encryption_key bucket_replication_key create
```
