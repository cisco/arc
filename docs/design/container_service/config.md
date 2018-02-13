# Container Service Configuration

The container service feature introduces the "container_service" section into arc datacenter configuration.
The "container_service" object is a peer with other top level objects such as the "datacenter", "database_service", "dns" and "notifications"
objects.


## container_service

The "container_service" section is a json object that contains two elements, the "provider" element and the "containers" element.
The "provider" element is a json object that defines the cloud provider data. The "containers" element is a json array of container instances.

```
  "container_service": {
    "name": "my-container-service",
    "provider": {
      ...
    }
  }
```

## provider

The provider element contains two elements, "vendor" and "data". The "vendor" element is a string indicating the cloud vendor to use.
Currently only the "aws" and "mock" vendors are supported. The "data" element is a json object comprised of key value pairs.
These pairs differ depending on the vendor.


### aws vendor

For an aws vendor, the following values are required:

- **account** (string) _required_: The aws account name
- **region**  (string) _required_: The aws region that this container service where this container service will reside

```json
    "provider": {
      "vendor": "aws",
      "data": {
        "account": "example-integration",
        "region":  "us-east-1"
      }
    }
```

### mock vendor

For a mock vendor, the data section is not required.

```json
  "container_service": {
    "provider": { "vendor": "mock" }
  }
```
