variable "resource_group" {
  description = "The name of the resource group in which to create the virtual network."
}

variable "location" {
  description = "The location/region where the virtual network is created. Changing this forces a new resource to be created."
  default     = "southcentralus"
}

variable "vault_name" {
  description = "The name of the vault resource"
}

variable "sku" {
  description = "SKU for the vault; allowed values are 'standard' and 'premium'."
  default     = "standard"
}

variable "tenant_id" {
  description = "Tenant Id of the subscription. Get using Get-AzureRmSubscription cmdlet or Get Subscription API"
}

variable "object_id" {
   description = "Object Id of the AD user. Get using Get-AzureRmADUser or Get-AzureRmADServicePrincipal cmdlets"
}

variable "key_permissions" {
  description = "Permissions to keys in the vault. Valid values are: all, create, import, update, get, list, delete, backup, restore, encrypt, decrypt, wrapkey, unwrapkey, sign, and verify."
  default     = "all"
}

variable "secret_permissions" {
  description = "Permissions to secrets in the vault. Valid values are: all, get, set, list, and delete."
  default     = "get"
}

variable "enabled_for_deployment" {
  description = "Specifies if the vault is enabled for a VM deployment"
  default     = false
}

variable "enabled_for_disk_encryption" {
  description = "Specifies if the azure platform has access to the vault for enabling disk encryption scenarios."
  default     = false
}

variable "enabled_for_template_deployment" {
  description = "Specifies whether Azure Resource Manager is permitted to retrieve secrets from the key vault."
  default     = false
}
