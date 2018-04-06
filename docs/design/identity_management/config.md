# Identity Management Configuration

The identity_management feature introduces the "identity_management" section to the amp account configuration. The "identity_management" object is peer to the other top level json objects like "storage" and "provider".

## identity management

The "identity_management" section is a json object that contains three elements, the "region" element, the "roles" and the "policies" element. The "region" element is a string, the "roles" element is a json array of roles, and the "policies" element is a json array of policies.

```
"identity_management" : {
  "region": "example-region",
  "roles": [
  ],
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

## roles
Since the identity management can have multiple roles, the roles element is an array of role elements.

### role

The role element represents a single policy in identity management. It has the following elements.
- **role** (string) _required_: This is the name of the role to be used to identify the run time role.
- **description** (string) _optional_: This is a description of the role and is checked during an audit.
- **instance_profile** (string) _optional_: This is needed if an instance needs to have a role. It is suggested that this is the same name as the role. There can only be one role per instance_profile.
- **trust_relationship** (string) _required_: This is a policy document that allows certain provider entities access to the role. This is the name of a json file that contains the policy_document located in "/etc/arc/policies/trust_relationships".
- **policies** (array) _required_: This is an array of strings with the name of policies to look up on the provider to attach to the role.

```
  {
    "role": "role_name",
    "description": "This is a sample role",
    "instance_profile": "role_name",
    "trust_relationship": "role_name_trust_relationship",
    "policies": [
      "policy_name",
      "replication_policy",
      "encryption_policy"
    ]
  }
```

## Example identity_management configuration
Here is an example of an account's identity management that has one policy "replication_policy", using AWS IAM policies with the policy document in "/etc/arc/policies/IAM_policies/replication_policy.json" and one role "replication_role" with the trust_relationship in "/etc/arc/policies/trust_relationships/replicatoin_role.json".

```
{
  "identity_management": {
    "region": "us-east-1",
    "roles": [
      {
        "role": "replication_role",
        "description": "policy to give resources CRR",
        "instance_profile": "replication_role",
        "trust_relationship": "replication_role",
        "policies": [
          "replication_policy"
        ]
      }
    ],
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
