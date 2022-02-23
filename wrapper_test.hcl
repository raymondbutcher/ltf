arrange "standard" {
  files = {
    "dev/dev.auto.tfvars" = "x = 1"
    "dev/dev.tfbackend"   = ""
    "main.tf"             = ""
  }

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
      TF_CLI_ARGS_init = "-backend-config=dev/dev.tfbackend"
      TF_DATA_DIR      = "dev/.terraform"
      TF_VAR_x         = "1"
    }
  }
}

arrange "fmt" {
  files = {
    "main.tf"                 = ""
    "subdir/terraform.tfvars" = "x = 1"
  }

  act "fmt" {
    cwd = "subdir"
    cmd = "ltf fmt terraform.tfvars"
  }

  assert "fmt bypasses ltf" {
    cmd = "terraform fmt terraform.tfvars"
    env = {
      TF_DATA_DIR = ""
      TF_VAR_x    = ""
    }
  }
}

arrange "variables" {
  files = {
    "live/blue/a.auto.tfvars" = <<-EOF
      x = "a"
    EOF
    "live/blue/d.auto.tfvars" = <<-EOF
      x = "d" # takes precedence in deepest directory
    EOF
    "live/b.auto.tfvars"      = <<-EOF
      x = "b"
      y = "b"
    EOF
    "live/e.auto.tfvars"      = <<-EOF
      x = "e"
      y = "e" # takes precedence in deepest directory
    EOF
    "live/terraform.tfvars"   = <<-EOF
      x = "terraform"
      y = "terraform"
    EOF
    "c.auto.tfvars"           = <<-EOF
      z = "c"
    EOF
    "f.auto.tfvars"           = <<-EOF
      z = "f" # takes precedence in config directory and trying to override in subdirectories would lead to and error
    EOF
    "main.tf"                 = <<-EOF
      variable "def" {
        default = "main"
      }
    EOF
  }

  act "plan" {
    cwd = "live/blue"
    cmd = "ltf plan"
  }

  assert "args" {
    cmd = "terraform -chdir=../.. plan"
    env = {
      TF_DATA_DIR       = "live/blue/.terraform"
      TF_CLI_ARGS_init  = ""
      TF_CLI_ARGS_plan  = ""
      TF_CLI_ARGS_apply = ""
      TF_VAR_x          = "d"
      TF_VAR_y          = "e"
      TF_VAR_z          = "f"
      TF_VAR_def        = "main"
    }
  }
}
