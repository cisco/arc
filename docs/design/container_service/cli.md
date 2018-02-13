# Container Service CLI Design

Two new commands will be added to the arc cli to provide management of containers with a datacenter.


To run an arc command against all containers in a datacenter use the form

```shell
arc [dc_name] container [command]
```

The commands available to the container service are

- **create**:  Create the container service.
- **destroy**: Destroy the container service.
- **update**:  Update the container service.
- **audit**:   Audits the container service.
- **info**:    Provides run time information for the container service.
- **config**:  Provides configuration information for the container service.
- **help**:    Provides help with the container service command.


For example, to run an audit command against the container service in the fictional "example-integration" datacenter, run the command

```shell
arc example-integration container audit
```

