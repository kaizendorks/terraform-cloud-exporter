terraform {
  cloud {
    workspaces {
      name = "example-error"
    }
  }
}

resource "terraform_data" "always_fails" {
  lifecycle {
    postcondition {
      condition     = self.output == "never"
      error_message = "Intentional failure for exporter testing."
    }
  }
}
