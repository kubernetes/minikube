variable "minikube_subscription_id" {
  description = "Identifier (GUID) of Azure subscription containing the resource group."
  type        = string
}

variable "minikube_resource_group" {
  description = "Name of Azure resource group where Packer will run Azure Image Builder and where is the Shared Image Gallery (SIG)."
  type        = string
}

variable "minikube_shared_image_gallery" {
  description = "Name of Azure Shared Image Gallery (SIG) to publish the VM image."
  type        = string
}

variable "vm_image_name" {
  description = "Name of the image build to be publish in the Shared Image Gallery (SIG) and referenced to create VM from it."
  type        = string
}

variable "vm_image_version" {
  description = "Version of the image build to be publish in the Shared Image Gallery (SIG) and referenced to create VM from it."
  type        = string
}

variable "vm_admin_username" {
  type = string
}

variable "vm_admin_password" {
  type = string
}

variable "vm_admin_ssh_public_key" {
  description = "Content of SSH public key to be deployed to VM image for admin authentication (base64-encoded PEM)."
  type        = string
}
