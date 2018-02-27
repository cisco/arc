# Key Management Configuration

The key_management feature introduces the "key_management" section to the amp account configuration. The "key_management" object is peer to the other top level json objects like "storage" and "provider".

## key management

The "key_management" section is a json object that contains two elements, the "region" element and the "encryption_keys" element. The "region" element is a string and the "encryption_keys" element is a json array of encryption keys.

```
"key_management" : {
  "region": "example-region",
  "encryption_keys": [
    ...
  ]
}
```

## encryption keys
Since the key management can have multiple encryption keys, the encryption_keys element is an array of encryption key elements.

### encryption key

The encryption key element represents a single encryption key in key management. It has the following elements.
- **encryption_key** (string) _required_: This is the alias of the key used to find the run time key.
- **deletion_pending_window** (int) _required_: This is the number of days that it will take to delete the key.

```
  {
    "encryption_key": "put_key_alias_here",
    "deletion_pending_window": 7
  }
```

## Example key_management configuration
Here is an example of an account's key management that has one key "bucket_replication_key", using AWS KMS encryption keys.

```
{
  "key_management": {
    "region": "us-east-1",
    "encryption_keys": [
      {
        "encryption_key": "bucket_replication_key",
        "deletiong_pending_window": 7
      }
    ]
  }
}
```
