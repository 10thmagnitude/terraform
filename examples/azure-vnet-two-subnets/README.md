# Virtual Network with two Subnets

<a href="http://armviz.io/#/?load=https%3A%2F%2Fraw.githubusercontent.com%2FAzure%2Fazure-quickstart-templates%2Fmaster%2F101-vnet-two-subnets%2Fazuredeploy.json" target="_blank">
    <img src="http://armviz.io/visualizebutton.png"/>
</a>

This template allows you to create a Virtual Network with two subnets.

## main.tf
The `main.tf` file contains the actual resources that will be deployed. It also contains the Azure Resource Group definition and any defined variables. 

## outputs.tf
This data is outputted when `terraform apply` is called, and can be queried using the `terraform output` command.

## provider.tf
Azure requires that an application is added to Azure Active Directory to generate the `client_id`, `client_secret`, and `tenant_id` needed by Terraform (`subscription_id` can be recovered from your Azure account details). Please go [here](https://www.terraform.io/docs/providers/azurerm/) for full instructions on how to create this to populate your `provider.tf` file.

## terraform.tfvars
If a `terraform.tfvars` file is present in the current directory, Terraform automatically loads it to populate variables. We don't recommend saving usernames and password to version control, but you can create a local secret variables file and use `-var-file` to load it.

## variables.tf
The `variables.tf` file contains all of the input parameters that the user can specify when deploying this Terraform template.