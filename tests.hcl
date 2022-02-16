arrange {
  files = [
    "dev/dev.auto.tfvars",
    "dev/dev.tfbackend",
    "main.tf"
  ]

  act {
    cwd = "dev"
    cmd = "ltf init"

    assert {
      cmd = "terraform -chdir=.. init"
    }
  }

  act {
    cwd = "dev"
    cmd = "ltf plan -target=random_id.this"

    assert {
      cmd = "terraform -chdir=.. plan -target=random_id.this"
    }
  }

  assert {
    env = {
      TF_DATA_DIR       = "dev/.terraform"
      TF_CLI_ARGS_init  = "-backend-config=dev/dev.tfbackend"
      TF_CLI_ARGS_plan  = "-var-file=dev/dev.auto.tfvars"
      TF_CLI_ARGS_apply = "-var-file=dev/dev.auto.tfvars"
    }
  }
}
