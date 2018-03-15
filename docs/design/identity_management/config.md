# Identity Management Configuration

The key_management feature introduces the "key_management" section to the amp account configuration. The "key_management" object is peer to the other top level json objects like "storage" and "provider".

## key management

The "identity_management" section is a json object that contains two elements, the "region" element and the "policies" element. The "region" element is a string and the "policies" element is a json array of policies.

```
"identity_management" : {
  "region": "example-region",
  "policies": [
    ...
  ]
}
```

## policies
Since the identity management can have multiple policies, the policies element is an array of policy elements.

### policy

The policy element represents a single policy in identity management. It has the following elements.
- **policy** (string) _required_: This is the name of the policy to be used to identify the run time policy.
- **description** (string) _optional_: This is a decription of the policy and is checked during an audit.
- **policy_document** (string) _required_: This is the name of the a separate json file that contains the policy configuration. This json file is located under "/etc/arc/policies/IAM_policies".

```
  {
    "policy": "policy_name",
    "description": "This is a sample policy",
    "policy_document": "policy_name"
  }
```

## Example identity_management configuration
Here is an example of an account's identity management that has one policy "replication_policy", using AWS IAM policies with the policy document in "/etc/arc/policies/IAM_policies/replication_policy.json".

```
{
  "identity_management": {
    "region": "us-east-1",
    "policies": [
      {
        "policy": "replication_policy",
        "description": "This policy enables CRR",
        "policy_document": "replication_policy",
      }
    ]
  }
}
```
