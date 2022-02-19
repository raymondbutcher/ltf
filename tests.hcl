arrange "standard" {
  files = [
    "dev/dev.auto.tfvars",
    "dev/dev.tfbackend",
    "main.tf"
  ]

  act "init" {
    cwd = "dev"
    cmd = "ltf init"

    assert "cmd" {
      cmd = "terraform -chdir=.. init"
    }
  }

  act "plan" {
    cwd = "dev"
    cmd = "ltf plan -target=random_id.this"

    assert "cmd" {
      cmd = "terraform -chdir=.. plan -target=random_id.this"
    }
  }

  assert "env" {
    env = {
      TF_DATA_DIR       = "dev/.terraform"
      TF_CLI_ARGS_init  = "-backend-config=dev/dev.tfbackend"
      TF_CLI_ARGS_plan  = "-var-file=dev/dev.auto.tfvars"
      TF_CLI_ARGS_apply = "-var-file=dev/dev.auto.tfvars"
    }
  }
}

arrange "fmt" {
  files = [
    "main.tf",
    "subdir/terraform.tfvars",
  ]

  act "fmt" {
    cwd = "subdir"
    cmd = "ltf fmt terraform.tfvars"
  }

  assert "env" {
    cmd = "terraform fmt terraform.tfvars"
    env = {
      TF_DATA_DIR       = ""
      TF_CLI_ARGS_init  = ""
      TF_CLI_ARGS_plan  = ""
      TF_CLI_ARGS_apply = ""
    }
  }
}

arrange "variables" {
  files = [
    "live/blue/a.auto.tfvars",
    "live/blue/d.auto.tfvars",
    "live/b.auto.tfvars",
    "live/e.auto.tfvars",
    "live/terraform.tfvars",
    "c.auto.tfvars",
    "f.auto.tfvars",
    "main.tf"
  ]

  act "plan" {
    cwd = "live/blue"
    cmd = "ltf plan"
  }

  assert "args" {
    cmd = "terraform -chdir=../.. plan"
    env = {
      TF_DATA_DIR       = "live/blue/.terraform"
      TF_CLI_ARGS_init  = ""
      TF_CLI_ARGS_plan  = "-var-file=live/terraform.tfvars -var-file=live/b.auto.tfvars -var-file=live/e.auto.tfvars -var-file=live/blue/a.auto.tfvars -var-file=live/blue/d.auto.tfvars"
      TF_CLI_ARGS_apply = "-var-file=live/terraform.tfvars -var-file=live/b.auto.tfvars -var-file=live/e.auto.tfvars -var-file=live/blue/a.auto.tfvars -var-file=live/blue/d.auto.tfvars"
    }
  }
}
