output "env" {
  value = [
    {
      name  = "REDSHIFT_USER"
      value = local.username
    },
    {
      name  = "REDSHIFT_DB"
      value = local.database_name
    }
  ]
}

output "secrets" {
  value = [
    {
      name  = "REDSHIFT_PASSWORD"
      value = local.password
    },
    {
      name  = "REDSHIFT_URL"
      value = "redshift://${urlencode(local.username)}:${local.password}@${local.db_endpoint}/${urlencode(local.database_name)}"
    }
  ]
}