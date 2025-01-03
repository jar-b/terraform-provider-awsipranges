data "awsipranges_ranges" "example" {
  filters = [
    {
      type   = "ip"
      values = ["3.5.12.4"]
    },
    {
      type   = "service"
      values = ["S3"]
    },
  ]
}
