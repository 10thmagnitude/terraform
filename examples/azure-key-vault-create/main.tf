resource "azurerm_resource_group" "rg" {
  name     = "${var.resource_group}"
  location = "${var.location}"
}

resource "azurerm_key_vault" "vault" {
  name                = "${var.vault_name}"
  location            = "${var.location}"
  resource_group_name = "${azurerm_resource_group.rg.name}"

  sku {
    name = "standard"
  }

  tenant_id = "${var.tenant_id}"

  access_policy {
    tenant_id = "${var.tenant_id}"
    object_id = "${var.object_id}"

    key_permissions = [
      "all",
    ]

    secret_permissions = [
      "get",
    ]
  }

  enabled_for_deployment          = "${var.enabled_for_deployment}"
  enabled_for_disk_encryption     = "${var.enabled_for_disk_encryption}"
  enabled_for_template_deployment = "${var.enabled_for_template_deployment}"
}
