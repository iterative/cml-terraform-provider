variables {
  gcp_build_region            = "us-west1-b"
  gcp_build_instance          = "custom-8-16384"
  gcp_build_accelerator_type  = "nvidia-tesla-k80"
  gcp_build_accelerator_count = 1
  gcp_build_ubuntu_image      = "ubuntu-1804-lts"
}

variables {
  gcp_project = "cml-gcp-test"
}

variable "gcp_post_processor" {
  value = <<-END
    cat manifest.json | jq --raw-output '.builds | .[].artifact_id' | while read image; do
      gcloud compute images add-iam-policy-binding "$image" \
        --member='allAuthenticatedUsers' --role='roles/compute.imageUser'
    done
    for family in "${var.image_name}"{,-test}; do
      gcloud compute images list --format=json --filter="family=$family" --sort-by=creationTimestamp |
      jq --raw-output '.[:-1] | .[].name' |
      while read image; do gcloud compute images delete "$image"; done
    done
  END
}

locals {
  tags = {
    environment = var.test ? "test" : "production"
  }
  release_regions = ["us"]
}

source "googlecompute" "source" {
  image_family            = var.test ? "${var.image_name}-test" : var.image_name
  image_name              = "${var.image_name}-{{timestamp}}"
  image_description       = var.image_description
  image_storage_locations = local.release_regions
  image_labels            = local.tags

  zone         = var.gcp_build_region
  machine_type = var.gcp_build_instance

  accelerator_count = var.gcp_build_accelerator_count
  accelerator_type = join("/", [
    "projects",
    var.gcp_project,
    "zones",
    var.gcp_build_region,
    "acceleratorTypes",
    var.gcp_build_accelerator_type,
  ])

  on_host_maintenance = "TERMINATE"

  source_image_family = "${var.gcp_build_ubuntu_image}"
  ssh_username        = "ubuntu"

  project_id = var.gcp_project
}
