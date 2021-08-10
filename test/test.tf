provider "grackdb" {
  api_url = "http://localhost:8081/query"
}

data "grackdb_current_user" "me" {}

resource "grackdb_user" "tf_user" {
    username = "Terraform User"
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
