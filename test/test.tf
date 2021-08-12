provider "grackdb" {
  api_url = "http://localhost:8081/query"
}

data "grackdb_current_user" "me" {}

resource "grackdb_user" "tf_user" {
  username = "Terraform User"
}

resource "grackdb_discord_account" "tf_discord" {
  discord_id    = "11111111111111112"
  username      = "Terraform"
  discriminator = "0001"
  owner         = grackdb_user.tf_user.id
}

output "id" {
  value = grackdb_user.tf_user.id
}

output "username" {
  value = grackdb_user.tf_user.username
}

output "avatar_url" {
  value = grackdb_user.tf_user.avatar_url
}
