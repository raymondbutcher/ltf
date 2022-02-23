// arrange "standard" {
//   files = {
//     "dev/dev.auto.tfvars" = "x = 1"
//     "dev/dev.tfbackend"   = ""
//     "main.tf"             = ""
//   }

//   act "init" {
//     cwd = "dev"
//     cmd = "ltf init"

//     assert "cmd" {
//       cmd = "terraform -chdir=.. init"
//     }
//   }

//   act "plan" {
//     cwd = "dev"
//     cmd = "ltf plan -target=random_id.this"

//     assert "cmd" {
//       cmd = "terraform -chdir=.. plan -target=random_id.this"
//     }
//   }

//   assert "env" {
//     env = {
//       TF_CLI_ARGS_init = "-backend-config=dev/dev.tfbackend"
//       TF_DATA_DIR      = "dev/.terraform"
//       TF_VAR_x         = "1"
//     }
//   }
// }

// arrange "fmt" {
//   files = {
//     "main.tf"                 = ""
//     "subdir/terraform.tfvars" = "x = 1"
//   }

//   act "fmt" {
//     cwd = "subdir"
//     cmd = "ltf fmt terraform.tfvars"
//   }

//   assert "fmt bypasses ltf" {
//     cmd = "terraform fmt terraform.tfvars"
//     env = {
//       TF_DATA_DIR = ""
//       TF_VAR_x    = ""
//     }
//   }
// }

arrange "variables" {
  files = {
    "live/blue/a.auto.tfvars" = <<-EOF
      x = 1
      y = 1
    EOF
    "live/blue/d.auto.tfvars" = <<-EOF
      x = 4
      y = 4
    EOF
    "live/b.auto.tfvars"      = <<-EOF
      x = 2
      y = 2
    EOF
    "live/e.auto.tfvars"      = <<-EOF
      x = 5
      y = 5
    EOF
    "live/terraform.tfvars"   = <<-EOF
      x = "terraform"
      y = "terraform"
    EOF
    "c.auto.tfvars"           = <<-EOF
      z = 3
    EOF
    "f.auto.tfvars"           = <<-EOF
      z = 6
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
      TF_VAR_x          = 4
      TF_VAR_y          = 4
      TF_VAR_z          = 6
      TF_VAR_def          = "main"
    }
  }
}

// arrange "vars" {
//   files = {
//     "dev/dev.auto.tfvars" = "x = 1"
//     "main.tf"             = <<-EOF
//       variable "x" {
//         type = number
//       }
//       variable "y" {
//         default = 34
//       }
//       variable "z" {
//         type = number
//       }
//     EOF
//     "bad.auto.tfvars"     = "z = 2"
//   }

//   act "plan" {
//     cwd = "dev"
//     cmd = "ltf init"
//   }

//   assert "env" {
//     env = {
//       TF_DATA_DIR       = "dev/.terraform"
//       TF_CLI_ARGS_init  = ""
//       TF_CLI_ARGS_plan  = ""
//       TF_CLI_ARGS_apply = ""
//       TF_VAR_x          = 1
//       TF_VAR_y          = 34
//       TF_VAR_z          = 2
//     }
//   }
// }
