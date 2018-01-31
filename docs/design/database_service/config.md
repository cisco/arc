# Database Service Configuration

The database service feature introduces the "database_service" section into arc datacenter configuration.
The "database_service" object is a peer with other top level objects such as the "datacenter", "dns" and "notifications"
objects.


## database_service

The "database_service" section is a json object that contains two elements, the "provider" element and the "databases" element.
The "provider" element is a json object that defines the cloud provider data. The "databases" element is a json array of database instances.

```
  "database_service": {
    "provider": {
      ...
    },

    "databases": [
      ...
    ]
  }
```

## provider

The provider element contains two elements, "vendor" and "data". The "vendor" element is a string indicating the cloud vendor to use.
Currently only the "aws" and "mock" vendors are supported. The "data" element is a json oject comprised of key value pairs.
These pairs differ depending on the vendor.


### aws vendor

For an aws vendor, the following value are required:

- **account** (string) _required_: The aws account name
- **region**  (string) _required_: The aws region that this database service where this database service will reside

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
  "database": {
    "provider": { "vendor": "mock" }
  }
```

## databases

Since a datacenter can most multi database instances, the database element is an array of database elements.

### database

The database element represent a single database instance in the datacenter. It has the following elements

- **database**        (string)  _required_: This is the name of the database instance. This name can be used on the cli to address this database instance.
- **engine**          (string)  _required_: The name of the database engine to use for this instance.
- **version**         (string)  _optional_: The version of the database engine to use for this instance.
- **type**            (string)  _required_: Similar to the pod type, this specifies the db instance class type. ([AWS DB Instance Types](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html))
- **port**            (integer) _optional_: The port the engine will uses for connections.
- **subnet_group**    (string)  _required_: The subnet group that this database instance will use.
- **security_groups** (string)  _required_: The security groups that are applied to the database instance.
- **storage**         (object)  _optional_:      The storage allocated to the database instance.
  - **type**            (string):             The type of storage.
  - **size**            (integer):            The size of the storage in GiB.
- **master**          (object)  _optional_
  - **username**:       (string):             The name of the master user.
  - **password**:       (string):             The password of the master user.

```json
      {
        "database":        "contacts",
        "engine":          "postgress",
        "version":         "9.6.5",
        "type":            "db.m4.large",
        "port":            5432,
        "subnet_group":    "postgres_subnet",
        "security_groups": [ "postgres_secgroup" ],
        "storage":         { "type": "gp2", "size": 50 },
        "master":          { "username": "myuser", "password": "mypasswd" }
      }
```


## Example database_service configuration

Here is an example of a simple postgres database using the AWS RDS service. There is a single database instance called "contacts".

```json
  "database_service": {
    "provider": {
      "vendor": "aws",
      "data": {
        "account": "example-integration",
        "region":  "us-east-1"
      }
    },
    "databases": [
      {
        "database":        "contacts",
        "engine":          "postgress",
        "version":         "9.6.5",
        "type":            "db.m4.large",
        "port":            5432,
        "subnet_group":    "postgres_subnet",
        "security_groups": [ "postgres_secgroup" ],
        "storage":         { "type": "gp2", "size": 50 },
        "master":          { "username": "myuser", "password": "mypasswd" }
      }
    ]
  },
```
