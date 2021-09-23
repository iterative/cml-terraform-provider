variables {
  aws_build_region       = "us-west-1"
  aws_build_instance     = "g2.2xlarge"
  aws_build_ubuntu_image = "*ubuntu-*-18.04-amd64-server-*"
}

variables {
  aws_role_session_name = "cml-packer-session"
  aws_role_arn          = "arn:aws:iam::260760892802:role/cml-packer"
  aws_security_group_id = "sg-0b7df7d9f902ca7ec"
  aws_subnet_id         = "subnet-09fca08419c2f0575"
}

locals {
  tags = {
    ManagedBy   = "packer"
    Name        = var.image_name
    Environment = var.test ? "test" : "prod"
  }
  release_regions = [
    "af-south-1",
    "ap-east-1",
    "ap-northeast-1",
    "ap-northeast-2",
    "ap-northeast-3",
    "ap-south-1",
    "ap-southeast-1",
    "ap-southeast-2",
    "ca-central-1",
    "eu-central-1",
    "eu-north-1",
    "eu-south-1",
    "eu-west-1",
    "eu-west-2",
    "eu-west-3",
    "me-south-1",
    "sa-east-1",
    "us-east-1",
    "us-east-2",
    "us-west-1",
    "us-west-2"
  ]
}

data "amazon-ami" "ubuntu" {
  region      = var.aws_build_region
  owners      = ["099720109477"]
  most_recent = true

  filters = {
    name                = "ubuntu/images/${var.aws_build_ubuntu_image}"
    root-device-type    = "ebs"
    virtualization-type = "hvm"
  }

  assume_role {
    role_arn     = var.aws_role_arn
    session_name = var.aws_role_session_name
  }
}

source "amazon-ebs" "source" {
  ami_groups      = ["all"]
  ami_name        = var.test ? var.image_name : "${var.image_name}-test"
  ami_description = var.image_description
  ami_regions     = var.test ? [local.aws_build_region] : local.release_regions

  region        = var.aws_build_region
  instance_type = var.aws_build_instance

  source_ami    = data.amazon-ami.ubuntu.id
  ssh_username  = "ubuntu"

  security_group_id = var.aws_security_group_id
  subnet_id         = var.aws_subnet_id

  force_delete_snapshot = true
  force_deregister      = true

  tags            = local.tags
  run_tags        = local.tags
  run_volume_tags = local.tags

  assume_role {
    role_arn     = var.aws_role_arn
    session_name = var.aws_role_session_name
  }
}
