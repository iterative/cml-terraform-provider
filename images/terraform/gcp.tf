data "google_organization" "iterative" {
  domain = "iterative.ai"
}

resource "google_project" "images" {
  org_id     = google_organization.iterative.org_id
  project_id = "cml-images"
  name       = "cml-images"
}

resource "google_service_account" "packer" {
  project = google_project.images.id

  account_id   = "service-account-id"
  display_name = "Service Account"
}

resource "google_service_usage_consumer_quota_override" "override" {
  provider       = google-beta
  project        = google_project.my_project.project_id
  service        = "compute.googleapis.com"
  metric         = "compute.googleapis.com%252Fgpus_all_regions"
  limit          = "GPUS-ALL-REGIONS-per-project"
  override_value = "1"
  force          = true
}
