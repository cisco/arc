# Getting Started

## AWS Configuration

This assumes that you already have an account with AWS.

### Bootstrapping IAM

TBD

### Configuration and Credential File Setup

You need to sign into the [AWS console](https://console.aws.amazon.com) and navigate to 
[Idenity and Access Management > Users](https://console.aws.amazon.com/iam/home?region=us-east-1#users).
Then navigate to your user name. From this page you are able to manage your AWS access keys. You 
should have permission to create, delete and deactivate your access keys. An access key will be required to 
setup the configuration and credentials files.

In order to communicate with AWS APIs you need to setup the config and credential files. 

See: [AWS Command Line Interface: Configuration and Credential Files](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files)

#### ~/.aws/config

If you are working with a multi-account setup, perhaps one account for integration and one account for production,
you want to have multiple profiles in your config file.

```shell
[default]
region=us-east-1
output=json

[profile example-integration]
region=us-east-2

[profile example-production]
region=us-east-1
output=json

```

#### ~/.aws/credentials

This example follows the lead from the above config example, where there are integration and production accounts.

```
[default]
aws_access_key_id = <access key of for default account>
aws_secret_access_key = <secret access key>

[example-integration]
aws_access_key_id = <access key of for example-integration account>
aws_secret_access_key = <example-integration secret access key>

[example-production]
aws_access_key_id = <access key of for example-production account>
aws_secret_access_key = <example-production secret access key>
