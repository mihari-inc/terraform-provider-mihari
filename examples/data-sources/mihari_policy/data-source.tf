data "mihari_policy" "example" {
  id = "existing-policy-uuid"
}

output "policy_name" {
  value = data.mihari_policy.example.name
}
