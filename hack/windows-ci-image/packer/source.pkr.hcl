source "azure-arm" "minikube-ci-windows-11" {
  subscription_id    = var.minikube_subscription_id
  use_azure_cli_auth = true

  # TODO: Does SIG-Windows grant us permission to allow Packer create/delete temporary resource group for build?
  build_resource_group_name = var.minikube_resource_group

  image_publisher = "MicrosoftWindowsDesktop"
  image_offer     = "Windows-11"
  image_sku       = "win11-25h2-pro"

  vm_size        = "Standard_D2s_v3"
  os_type        = "Windows"
  communicator   = "winrm"
  winrm_use_ssl  = true
  winrm_insecure = true
  winrm_timeout  = "5m"
  winrm_username = var.vm_admin_username
  winrm_password = var.vm_admin_password

  shared_image_gallery_destination {
    subscription            = var.minikube_subscription_id
    resource_group          = var.minikube_resource_group
    gallery_name            = var.minikube_shared_image_gallery
    image_name              = var.vm_image_name
    image_version           = var.vm_image_version
    use_shallow_replication = true # https://learn.microsoft.com/en-us/azure/virtual-machines/shared-image-galleries#shallow-replication
    specialized             = true # https://learn.microsoft.com/en-us/azure/virtual-machines/shared-image-galleries#generalized-and-specialized-images
  }
}