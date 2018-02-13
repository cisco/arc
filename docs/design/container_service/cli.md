# Container Service CLI Design

Two new commands will be added to the arc cli to provide management of containers with a datacenter.


To run an arc command against all containers in a datacenter use the form

```shell
arc [dc_name] container [command]
```

The commands available to all containers are

- **create**:  Create the container service.
- **destroy**: Destroy the tags for all container instances.
- **update**:  Update the tags for all container instances.
- **audit**:   Audits all container instances.
- **info**:    Provides run time information for the container service and all its instances.
- **config**:  Provides configuration information for the container service and all its instances.
- **help**:    Provides help with the container command.


For example, to run an audit command against the container service in the fictional "example-integration" datacenter, run the command

```shell
arc example-integration container audit
```

