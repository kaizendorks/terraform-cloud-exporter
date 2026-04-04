terraform {
  cloud {
    workspaces {
      name = "example-success"
    }
  }
}

resource "random_string" "example" {
  length  = 16
  special = false
}
