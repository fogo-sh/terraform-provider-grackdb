provider "grackdb" {
  api_url = "http://localhost:8081/query"
}

data "grackdb_current_user" "me" {}

output "id" {
  value = data.grackdb_current_user.me.id
}

output "username" {
  value = data.grackdb_current_user.me.username
}

output "avatar_url" {
  value = data.grackdb_current_user.me.avatar_url
}
