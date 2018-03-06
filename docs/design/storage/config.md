# Storage Configuration

The storage feature introduces the "storage" section into the amp account configuration. The "storage" object is peer to the other top level json objects like "provider" and "notifications".

## storage

The "storage" section is a json object that contains two elements, the "buckets" element and the "buckets_sets" element. The "buckets" element is a json array of buckets. The "bucket_sets" element is a json array of bucket sets. A bucket set is a list of buckets that have cross region replication enabled.

```
"storage" :{
  "buckets": [
    ...
   ],

  "bucket_sets": [
    ...
  ]
}
```

## buckets
Since the storage can have multiple buckets, the buckets element is an array of bucket elements.

### bucket

**NOTE** 
The elements role and destination are only used if the buckets are part of a bucket set.
For replication of encrypted buckets the elements "encryption_key" and "destination_encryption_key" are both required and must be in different regions.

The bucket element represents a single bucket in the storage. It has the following elements.
- **bucket** (string) _required_: This is the name of the bucket. This name can be used on the cli to address this bucket.
- **region** (string) _required_: This is the region that the bucket will be created on.
- **encrytpion_key** (string) _optional_: This is an alias to a key that is used to encrypt the bucket.
- **role** (string) _optional_: This is the role used for cross region bucket replication. This is required for bucket sets.
- **destination** (string) _optional_: This is the target bucket for replication. This is required for bucket sets.

```
  {
    "bucket": "put_bucket_name_here",
    "region": "put_bucket_region_here",
    "role": "my_Replication_Role",
    "destination": "target_bucket_for_replication"
  }
```
## bucket sets
Since the storage can have multiple bucket sets, the bucket_sets element in an array of bucket set elements.

### bucket set
The bucket set element is a collection of buckets that must be created and destroyed together for the purpose of cross region bucket replication for all buckets in the bucket set. It has the following elements.
- **bucket_set** (string) _required_: This is the name of the bucket set. This name can be used on the cli to address this bucket set.
- **buckets** (object) _required_: The buckets is an array of bucket objects that require their Role and Destination specified
  - **bucket** (string) _required_: This is the name of the bucket. This name can be used on the cli to address this bucket.
  - **region** (string) _required_: This is the region that the bucket will be created on.
  - **encrytpion_key** (string) _optional_: This is an alias to a key that is used to encrypt the bucket.
  - **role** (string) _required_: This is the role used for cross region bucket replication. This role must be the same for each bucket within the bucket set.
  - **destination** (string) _required_: This is the target bucket for replication.
  - **destination_encrytpion_key** (string) _optional_: This is an alias to a key that is used to encrypt the destination bucket. This encryption key needs to be in the same region as the destination bucket.


```
  {
    "bucket_set": "my_replcation_bucket_set",
    "buckets": [
      {
        "bucket": "my-replication-bucket-us-east-1",
        "region": "us-east-1",
        "role": "my-Replication-Role",
        "destination": "my-replication-bucket-us-east-2"
      },
      {
        "bucket": "my-replication-bucket-us-east-2",
        "region": "us-east-2",
        "role": "my-Replication-Role",
        "destination": "my-replication-bucket-us-east-1"
      },
    ]
  },
  {
    "bucket_set": "my_encrypted_bucket_set",
    "buckets": [
      {
        "bucket": "my_encrypted_bucket-us-east-1",
        "region": "us-east-1",
        "encryption_key": "example_key_alias-us-east-1",
        "role": "my-Encryption-Role",
        "destination": "my_encrypted_bucket-us-east-2",
        "destination_encryption_key": "example_key_alias-us-east-2"
      },
      {
        "bucket": "my_encrypted_bucket-us-east-2",
        "region": "us-east-2",
        "encryption_key": "example_key_alias-us-east-2",
        "role": "my-Encryption-Role",
        "destination": "my_encrypted_bucket-us-east-1",
        "destination_encryption_key": "example_key_alias-us-east-1"
      }
    ]
  }
```
## Example storage configuration
Here is an example of an account's storage that has one bucket "sample-bucket", one encrypted bucket "encrypted-bucket", one bucket set "sample-replication-bucket-set" that is not encrypted, and one encrypted bucket set "encrypted_bucket_set" using AWS S3 buckets.

```
{
  "storage": {
    "buckets": [
      {
        "bucket": "sample-bucket",
        "region": "us-east-1"
      },
      {
        "bucket": "encrypted-bucket",
        "region": "us-east-1",
        "encryption_key": "example_encryption_key-us-east-1"
      }
    ],

    "bucket_sets": [
      {
        "bucket_set": "sample-replication-bucket-set",
        "buckets": [
          {
            "bucket": "sample-replicated-bucket-us-east-1",
            "region": "us-east-1",
            "role": "S3_replication_Role_for_sample_bucket_set",
            "Destination": "sample-replicated-bucket-us-east-2"
          },
          {
            "bucket": "sample-replicated-bucket-us-east-2",
            "region": "us-east-2",
            "role": "S3_replication_Role_for_sample_bucket_set",
            "Destination": "sample-replicated-bucket-us-east-1"
          }
        ]
      },
      {
        "bucket_set": "encrypted_bucket_set",
        "buckets": [
          {
            "bucket": "encrypted_bucket-us-east-1",
            "region": "us-east-1",
            "encryption_key": "example_key_alias-us-east-1",
            "role": "my-Encryption-Role",
            "destination": "my_encrypted_bucket-us-east-2",
            "destination_encryption_key": "example_key_alias-us-east-2"
          },
          {
            "bucket": "encrypted_bucket-us-east-2",
            "region": "us-east-2",
            "encryption_key": "example_key_alias-us-east-2",
            "role": "my-Encryption-Role",
            "destination": "my_encrypted_bucket-us-east-1",
            "destination_encryption_key": "example_key_alias-us-east-1"
          }
        ]
      }
    ]
  }
}

```
