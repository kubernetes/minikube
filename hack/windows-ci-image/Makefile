# minikube-windows-image Makefile

MINIKUBE_AZ_SIG_STACK_DEPLOYMENT_NAME ?= "make-minikube-ci-sig" # Azure implementation detail
MINIKUBE_AZ_VNET_STACK_DEPLOYMENT_NAME ?= "make-minikube-ci-vnet" # Azure implementation detail
MINIKUBE_AZ_VM_STACK_DEPLOYMENT_NAME ?= "make-minikube-ci-vm" # Azure implementation detail

.DEFAULT_GOAL := help

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY:
preflight: ## Pre-flight checks for azure-* and packer-* targets
	$(if $(strip $(MINIKUBE_AZ_SUBSCRIPTION_ID)),,$(error MINIKUBE_AZ_SUBSCRIPTION_ID is not defined. Source env.sh.))
	$(if $(strip $(MINIKUBE_AZ_RESOURCE_GROUP)),,$(error MINIKUBE_AZ_RESOURCE_GROUP is not defined. Source env.sh.))
	$(if $(strip $(MINIKUBE_AZ_SIG_NAME)),,$(error MINIKUBE_AZ_SIG_NAME is not defined. Source env.sh.))
	$(if $(strip $(MINIKUBE_AZ_IMAGE_NAME)),,$(error MINIKUBE_AZ_IMAGE_NAME is not defined. Source env.sh.))
	$(if $(strip $(MINIKUBE_AZ_IMAGE_VERSION)),,$(error MINIKUBE_AZ_IMAGE_VERSION is not defined. Source env.sh.))
	$(if $(strip $(MINIKUBE_AZ_VM_ADMIN_USERNAME)),,$(error MINIKUBE_AZ_VM_ADMIN_USERNAME is not defined. Source env.sh.))
	$(if $(strip $(MINIKUBE_AZ_VM_ADMIN_PASSWORD)),,$(error MINIKUBE_AZ_VM_ADMIN_PASSWORD is not defined. Add to env.local.sh.))
	$(if $(strip $(MINIKUBE_AZ_VM_ADMIN_SSH_PUBLIC_KEY)),,$(error MINIKUBE_AZ_VM_ADMIN_SSH_PUBLIC_KEY is not defined. Add to env.local.sh.))
	@$(MAKE) printenv

.PHONY:
printenv: ## Dump project-specific environment variables
	@echo
	@printf "\033[0;32m%(%F %T)T Azure Minikube CI environment variables:\033[0m\n"
	@env | grep AZ | sort | sed "s|$(MINIKUBE_AZ_VM_ADMIN_PASSWORD)|*****|g"

##@ Azure resources inspection

.PHONY:
_az-header:
	@echo
	@printf "\033[0;32m%(%F %T)T Executing Azure CLI command:\033[0m\n"

.PHONY:
az-list-resources: | preflight _az-header ## List everything in resourcs group
	az resource list --resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" --output table

.PHONY:
az-list-deployments: | preflight _az-header ## List deployment stacks on Azure
	az deployment group list --verbose --resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" --output table
	@echo
	az stack group list --verbose --resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" --output table

##@ Azure Shared Image Gallery lifecycle

.PHONY:
sig-plan: | preflight _az-header ## Generate what-if plan of SIG stack deployment for inspection
	az deployment group what-if --verbose \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file sig.bicep \
		--parameters sig.bicepparam

.PHONY:
sig-deploy: | preflight _az-header ## Deploy SIG stack from Bicep templates to Azure
	az stack group create --verbose \
		--name "$(MINIKUBE_AZ_SIG_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file sig.bicep \
		--parameters sig.bicepparam \
		--action-on-unmanage deleteResources --deny-settings-mode None

.PHONY:
sig-validate: | preflight _az-header ## Validate SIG stack deployment
	az stack group validate --verbose \
		--name "$(MINIKUBE_AZ_SIG_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file sig.bicep \
		--parameters sig.bicepparam \
		--action-on-unmanage detachAll --deny-settings-mode none

.PHONY:
sig-undeploy: | preflight _az-header ## Delete SIG stack and its resources from Azure (prompts for confirmation)
	az stack group delete --verbose \
		--name "$(MINIKUBE_AZ_SIG_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--action-on-unmanage deleteResources

.PHONY:
sig-list-images: | preflight _az-header ## Lists image versions available in SIG
	az sig image-version list --verbose \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--gallery-name "$(MINIKUBE_AZ_SIG_NAME)" \
		--gallery-image-name "$(MINIKUBE_AZ_IMAGE_NAME)" \
		--query '[].{Version:name,PublishedDate:publishingProfile.publishedDate,Location:location,group:resourceGroup,state:provisioningState}' \
		--output table

##@ Azure Virtual Network + Network Security Group lifecycle (required by VM)

.PHONY:
vnet-plan: | preflight _az-header ## Generate what-if plan of Network stack deployment for inspection
	az deployment group what-if --verbose \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file vnet.bicep \
		--parameters vnet.bicepparam

.PHONY:
vnet-deploy: | preflight _az-header ## Deploy Network stack from Bicep templates to Azure
	az stack group create --verbose \
		--name "$(MINIKUBE_AZ_VNET_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file vnet.bicep \
		--parameters vnet.bicepparam \
		--action-on-unmanage deleteResources --deny-settings-mode None

.PHONY:
vnet-validate: | preflight _az-header ## Validate Network stack deployment
	az stack group validate --verbose \
		--name "$(MINIKUBE_AZ_VNET_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file vnet.bicep \
		--parameters vnet.bicepparam \
		--action-on-unmanage detachAll --deny-settings-mode none


.PHONY:
vnet-undeploy: | preflight _az-header ## Delete Network stack and its resources from Azure (prompts for confirmation)
	az stack group delete --verbose \
		--name "$(MINIKUBE_AZ_VNET_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--action-on-unmanage deleteResources

##@ Azure Virtual Machine lifecycle

.PHONY:
vm-plan: | preflight _az-header ## Generate what-if plan of VM stack deployment for inspection
	az deployment group what-if --verbose \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file vm.bicep \
		--parameters vm.bicepparam

.PHONY:
vm-deploy: | preflight _az-header ## Deploy VM stack from Bicep templates to Azure
	az stack group create --verbose \
		--name "$(MINIKUBE_AZ_VM_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file vm.bicep \
		--parameters vm.bicepparam \
		--action-on-unmanage deleteResources --deny-settings-mode None

.PHONY:
vm-validate: | preflight _az-header ## Validate VM stack deployment
	az stack group validate --verbose \
		--name "$(MINIKUBE_AZ_VM_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--template-file vm.bicep \
		--parameters vm.bicepparam \
		--action-on-unmanage detachAll --deny-settings-mode none

.PHONY:
vm-undeploy: | preflight _az-header ## Undeploy VM stack and its resources from Azure (sloow, prompts for confirmation)
	az stack group delete --verbose \
		--name "$(MINIKUBE_AZ_VM_STACK_DEPLOYMENT_NAME)" \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--subscription "$(MINIKUBE_AZ_SUBSCRIPTION_ID)" \
		--action-on-unmanage deleteResources

.PHONY:
vm-delete: | preflight _az-header ## Delete VM with fast direct delete operations (args: vm=<name> and optional rg=<name>; 10x faster than undeploy)
	./scripts/az-vm-delete.sh $(vm) $(rg)

.PHONY:
vm-fqdn: | preflight _az-header ## Lists FQDN of all VMs deployed on Azure
	az network public-ip list \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--query '[].{IP:ipAddress, Hostname:dnsSettings.fqdn}' \
		--output table

.PHONY:
vm-list: | preflight _az-header ## List VMs deployed on Azure
	az vm list --verbose \
		--resource-group "$(MINIKUBE_AZ_RESOURCE_GROUP)" \
		--show-details \
		--output table

##@ Packer pipeline building VM image

.PHONY:
_packer-header:
	@echo
	@printf "\033[0;32m%(%F %T)T Executing Packer command:\033[0m\n"

.PHONY:
_packer-generate-vars: | preflight
	@echo
	@printf "\033[0;32m%(%F %T)T Generating packer/azure.auto.pkrvars.hcl:\033[0m\n"
	@echo "# Azure Shared Image Gallery where Packer will  publish the VM image." > packer/azure.auto.pkrvars.hcl
	@echo "minikube_subscription_id=\"$(MINIKUBE_AZ_SUBSCRIPTION_ID)\"" >> packer/azure.auto.pkrvars.hcl
	@echo "minikube_resource_group=\"$(MINIKUBE_AZ_RESOURCE_GROUP)\"" >> packer/azure.auto.pkrvars.hcl
	@echo "minikube_shared_image_gallery=\"$(MINIKUBE_AZ_SIG_NAME)\"" >> packer/azure.auto.pkrvars.hcl
	@echo "vm_image_name=\"$(MINIKUBE_AZ_IMAGE_NAME)\"" >> packer/azure.auto.pkrvars.hcl
	@echo "vm_image_version=\"$(MINIKUBE_AZ_IMAGE_VERSION)\"" >> packer/azure.auto.pkrvars.hcl
	@echo "vm_admin_password=\"$(MINIKUBE_AZ_VM_ADMIN_PASSWORD)\"" >> packer/azure.auto.pkrvars.hcl
	@echo "vm_admin_username=\"$(MINIKUBE_AZ_VM_ADMIN_USERNAME)\"" >> packer/azure.auto.pkrvars.hcl
	@echo "vm_admin_ssh_public_key=\"$(MINIKUBE_AZ_VM_ADMIN_SSH_PUBLIC_KEY)\"" >> packer/azure.auto.pkrvars.hcl

.PHONY:
packer-fmt: _packer-header ## Run packer fmt
	packer fmt ./packer

.PHONY:
packer-init: _packer-generate-vars ## Run packer init preparing for VM image build 
	packer init ./packer

.PHONY:
packer-validate: packer-init ## Run packer validate
	cd ./packer && packer validate .

.PHONY:
packer-build-and-publish: packer-init ## Run packer build (checks shared image gallery exists, disallows image overwrite)
	cd ./packer && packer build -timestamp-ui -warn-on-undeclared-var .

.PHONY:
packer-build-and-overwrite: packer-init ## Run packer build (checks shared image gallery exists, overwrites existing image)
	cd ./packer && packer build -force -timestamp-ui -warn-on-undeclared-var .

##@ Bicep templates development

templates := $(shell find . -type f \( -name '*.bicep' -o -name '*.bicepparam' \) )

.PHONY:
_bicep-header:
	@echo
	@printf "\033[0;32m%(%F %T)T Executing Bicep command:\033[0m\n"

_bicep-fmt: _bicep-header $(templates:.bicep=.bicep.fmt) $(parameters:.bicepparam=.biceparam.fmt)
bicep-fmt: ## Format Bicep templates
	@$(MAKE) _bicep-fmt

_bicep-lint: _bicep-header $(templates:.bicep=.bicep.lint) $(parameters:.bicepparam=.biceparam.lint)
bicep-lint: ## Verify Bicep templates
	@$(MAKE) _bicep-lint

%.bicep.fmt %.bicepparam.fmt:
	AZURE_BICEP_CHECK_VERSION=False az bicep format --file $(basename $@) --verbose 2>&1

%.bicep.lint %.bicepparam.lint:
	AZURE_BICEP_CHECK_VERSION=False az bicep lint --file $(basename $@) --verbose 2>&1
