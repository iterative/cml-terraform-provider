packer {
  required_plugins {
    amazon = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/amazon"
    }
    azure = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/azure"
    }
    googlecompute = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/googlecompute"
    }
  }
}

variable "test" {
  type    = bool
  default = false
}

variables {
  image_name        = "iterative-cml"
  image_description = "CML (Continuous Machine Learning) Ubuntu 18.04"
}

build {
  sources = [
    "source.amazon-ebs.source",
    "source.azure-arm.source",
    "sources.googlecompute.source"
  ]

  provisioner "shell" {
   script = "${path.root}/../provisioner/setup.sh"
  }

  # provisioner "ansible" {
  #   playbook_file = "${path.root}/../ansible/playbook.yml"
  #   galaxy_file   = "${path.root}/../ansible/requirements.yml"
  # }

  post-processor "manifest" {
    output = "manifest.json"
    strip_path = true
  }

  post-processor "shell-local" {
    inline = [
      var.gcp_post_processor,
      var.azure_post_processor,
    ]
  }
}
