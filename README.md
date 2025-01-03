# Terraform Provider AWS IP Ranges

Terraform provider for working with public AWS IP range data.

This provider offers the same functionality as the [`aws_ip_ranges` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ip_ranges) in the [AWS provider](https://registry.terraform.io/providers/hashicorp/aws/latest), but with some marginal benefits such as a smaller binary size and optional caching to improve performance.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

This provider currently only exposes one data source for fetching and filtering IPv4 ranges published by AWS.
The ranges can be filtered by IP address, region, network border group, or service.
Example are included below.

### Filterting by IP Address

```terraform
data "awsipranges_ranges" "example" {
  filters = [
    {
      type   = "ip"
      values = ["3.5.12.4"]
    }
  ]
}
```

### Filtering by Region

```terraform
data "awsipranges_ranges" "example" {
  filters = [
    {
      type   = "region"
      values = ["us-east-1"]
    }
  ]
}
```

### Filtering by Service

```terraform
data "awsipranges_ranges" "example" {
  filters = [
    {
      type   = "service"
      values = ["S3"]
    }
  ]
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

```shell
make generate
```

In order to run the full suite of Acceptance tests, run `make testacc`.

```shell
make testacc
```
