# AWS Keymaster

[![Build Status](https://travis-ci.org/ryane/aws-keymaster.svg)](https://travis-ci.org/ryane/aws-keymaster)

A simple utility that allows you import a public key into all AWS regions with a single command.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [AWS Keymaster](#aws-keymaster)
    - [Building](#building)
    - [Configuring](#configuring)
    - [Running](#running)
        - [Import a public key into all regions](#import-a-public-key-into-all-regions)
        - [Delete a named keypair from all regions](#delete-a-named-keypair-from-all-regions)
        - [Dry Runs](#dry-runs)
        - [Running from Docker](#running-from-docker)
    - [License](#license)

<!-- markdown-toc end -->

## Building

Use the `Makefile` to build `aws-keymaster`:

```shell
make build
```

To build a docker container:

```shell
docker build -t aws-keymaster .
```

## Configuring

Before running `aws-keymaster`, you need to ensure that you have configured access to your AWS account. You can do so by using the [AWS CLI](https://aws.amazon.com/cli/) to [configure](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) your development machine. Alternatively, you can configure credentials by creating a file called `~/.aws/credentials` with contents that look something like this:

```
[default]
aws_access_key_id = AKID1234567890
aws_secret_access_key = MY-SECRET-KEY
```

Or, you can use environment variables to configure your credentials.

```
AWS_ACCESS_KEY_ID=AKID1234567890
AWS_SECRET_ACCESS_KEY=MY-SECRET-KEY
```

Amazon has a [blog post](http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs) with more information about how to configure your AWS credentials.

The credentials you use must be associated with an [IAM user](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/ec2-api-permissions.html) that has sufficient [permissions](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/ec2-api-permissions.html) to import keypairs in all regions.

## Running

```shell
Usage:
  aws-keymaster [command]

Available Commands:
  import      Imports a public key into all AWS regions
  delete      Deletes a keypair from all AWS regions
  version     Display the version of aws-keymaster

Flags:
      --dry-run[=false]: Checks whether you have the required permissions, without attempting the request
  -h, --help[=false]: help for aws-keymaster

Use "aws-keymaster [command] --help" for more information about a command.
```

### Import a public key into all regions

```shell
Imports a public key with the specified name and public key to all AWS regions

Usage:
  aws-keymaster import [name] [public key file] [flags]

Global Flags:
      --dry-run[=false]: Checks whether you have the required permissions, without attempting the request
```

To import a public key to all regions, you use the `import` command. It requires two arguments: the name of the key pair and the path to the public key on your file system.

```shell
./bin/aws-keymaster import mykey ~/.ssh/id_rsa.pub
eu-west-1:       Imported keypair 'mykey' - 57:bf:37:68:69:18:29:aa:4d:da:f7:1b:e6:28:4e:e8
ap-southeast-1:  Imported keypair 'mykey' - 57:bf:37:68:69:18:29:aa:4d:da:f7:1b:e6:28:4e:e8
...
```

If you do not pass in those arguments, `aws-keymaster` will prompt you for them.

```shell
./bin/aws-keymaster import
Key Name: mypubkey
Public key [/Users/ryan/.ssh/id_rsa.pub]:

eu-west-1:       Imported keypair 'mypubkey' - 57:bf:37:68:69:18:29:aa:4d:da:f7:1b:e6:28:4e:e8
...
```

### Delete a named keypair from all regions

```shell
Deletes a keypair with the specified name from all AWS regions

Usage:
  aws-keymaster delete [name] [flags]

Flags:
  -f, --force[=false]: Delete keypairs without prompting

Global Flags:
      --dry-run[=false]: Checks whether you have the required permissions, without attempting the request
```

To delete a keypair called `keypairname` from all regions, you can use the `delete` subcommand:

```shell
./bin/aws-keymaster delete keypairname
Are you sure you want to delete keypair 'testing'? (yes/no) [no]: yes
eu-west-1:       Deleted keypair 'keypairname'
ap-southeast-1:  Deleted keypair 'keypairname'
...
```

You can delete a keypair without prompting by using the `--force` flag:

```shell
./bin/aws-keymaster delete keypairname --force
```

### Dry Runs

For both the `import` and `delete` commands, you can use the `--dry-run` flag to confirm that your AWS credentials have the sufficient permissions to perform the operations:

```shell
./bin/aws-keymaster import mykey ~/.ssh/id_rsa.pub --dry-run
[Dry Run] eu-west-1:       Imported keypair 'mykey'
[Dry Run] ap-southeast-1:  Imported keypair 'mykey'
...
```

### Running from Docker

If you are running `aws-keymaster` from a docker container, you will likely need to use a volume mount in order to be able to specify a public key on the local file system. In addition, you may need to use environment variables to pass in your AWS credentials. The example below illustrates how to use the docker container to run the `import` command:

```shell
docker run --rm -it -v ~/.ssh/:/ssh -e "AWS_ACCESS_KEY_ID=AKID1234567890" -e "AWS_SECRET_ACCESS_KEY=MY-SECRET-KEY" ryane/aws-keymaster import dockertest /ssh/id_rsa.pub
```

## License

`aws-keymaster` is released under the Apache 2.0 license (see [LICENSE](LICENSE))
