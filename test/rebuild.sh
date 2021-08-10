#!/bin/sh

go build ../
cp ./terraform-provider-grackdb ~/.terraform.d/plugins/registry.terraform.io/nint8835/grackdb/0.1.0/darwin_amd64/
rm .terraform.lock.hcl
terraform init
