data "awsipranges_ranges" "example" {
  filters = [
    {
      type  = "ip"
      value = "3.5.12.4"
    }
  ]
}
