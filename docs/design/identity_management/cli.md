# Identity Management CLI Design

Two new commands will be added to the amp cli to provide management of encryption keys.


## Identity Management

To run an amp command against all children of identity_management in an account use the form

```shell
amp [account_name] identity_management [command] [flags]
```

The commands available to identity_management are

- **audit**: Audits the entirety of identity management.
- **info**: Provides run time information for identity management.
- **config**: Provides configuration information for the identity management.
- **help**: Provides help with the key_management command.


For example, to run an audit command against the identity_management in the fictional "example-integration" account, run the command

```shell
amp example-integration identity_management audit
```


## Named Policy

To run a command against a single policy in identity_management use the form

```shell
amp [account_name] policy [policy_name] [command] [flags]
```

where "policy_name" is the name of the policy.

The commands available to a policy are

- **create**: Creates the policy. If the policy already exists this command will do nothing.
- **destroy**: Deletes the policy. If the policy does not exist this command will do nothing.
- **audit**: Audits the policy. This will compare the configuration policy and report if the run time policy does not match.
- **info**: Provides run time information for the policy.
- **config**: Provides configuration information for the policy.
- **help**: Provides help with the named policy command.

For example, to create a policy with the name "replication_policy" in the "example-integration" account, run the command

```shell
amp example-integration policy replication_policy create
```


## Named Role

To run a command against a single role in identity_management use the form 

```shell
amp [account_name] role [role_name] [command] [flags]
```

where "role_name" is the name of the role.

The commands available to a role are

- **create**: Creates the role. If the role already exists this command will do nothing.
- **destroy**: Deletes the role. If the role does not exist this command will do nothing.
- **update**: Updates the description of the role as well as syncing the attached policies with the configured policies.
- **audit**: Audit the role. This will compare the configuration role and report if the run time role does not match.
- **info**: Provides run time information for the role.
- **config**: Provides configuration information for the role.
- **help**: Provides help with the named role command.

For example, to create a role with the name "replication_role" in the "example-integration" account, run the command

```shell
amp example-integration role replication_role create
```
