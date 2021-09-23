variables {
  azure_build_region       = "eastus"
  azure_build_instance     = "Standard_NC6"
  azure_build_ubuntu_image = "18.04-LTS"
}

variables {
  azure_resource_group  = "cml-packer-resource-group"
  azure_storage_account = "iterative"
}

variables {
  azure_client_id       = env("AZURE_CLIENT_ID")
  azure_client_secret   = env("AZURE_CLIENT_SECRET")
  azure_subscription_id = env("AZURE_SUBSCRIPTION_ID")
  azure_tenant_id       = env("AZURE_TENANT_ID")
}

locals {
  image_name = var.test ? "${var.image_name}-test" : var.image_name
}

locals {
  tags = {
    Env   = var.test ? "Test" : "Prod"
    Owner = "packer"
  }
}

source "azure-arm" "source" {
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  subscription_id = var.azure_subscription_id
  tenant_id       = var.azure_tenant_id

  capture_container_name = local.image_name
  capture_name_prefix    = local.image_name
  storage_account        = var.azure_storage_account
  resource_group_name    = var.azure_resource_group

  location = var.azure_build_region
  vm_size  = var.azure_build_instance

  os_type         = "Linux"
  image_offer     = "UbuntuServer"
  image_publisher = "Canonical"
  image_sku       = var.azure_build_ubuntu_image

  azure_tags = local.tags
}
