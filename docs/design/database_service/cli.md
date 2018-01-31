# Database Service CLI Design

Two new commands will be added to the arc cli to provide management of databases with a datacenter.


## All databases

To run an arc command against all databases in a datacenter use the form

```shell
arc [dc_name] database [command]
```

The commands available to all databases are

- **update**: Update the tags for all database instances.
- **audit**:  Audits all database instances.
- **info**:   Provides run time information for the database service and all its instances.
- **config**: Provides configuration information for the database service and all its instances.
- **help**:   Provides help with the database command.


For example, to run an audit command against all databases in the fictional "example-integration" datacenter, run the command

```shell
arc example-integration database audit
```


## Named database

To run a command against a single database instance in a datacenter use the form

```shell
arc [dc_name] database [database_name] [command]
```

where "database_name" is the name of the database instance.


The commands available to a database instance are

- **create**:  Creates the database instance. If the database already exists this command will do nothing.
- **destroy**: Destroyes the database instance. If the database does not exist the command will do nothing.
- **update**:  Updates the tags for this database instance. If the database does not exist the command will do nothing.
- **audit**:  Audit the database instance. The audit will compare the configuration of the database instance and report if the run time instance does not match.
- **info**:   Provides run time information for the database instance.
- **config**: Provides configuration information for the database instance.
- **help**:   Provides help with the named database command.


For example, to create a database named "contacts" in the "example-integration" datacenter, run the command

```shell
arc example-integration database contacts create
```



