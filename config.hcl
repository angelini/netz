global {
  log_directory      = "./logs"
  ingress_port       = 8080
  connection_timeout = "2s"
}

service "http" "api" {
  address             = "api.svc.local"
  local_port          = 5000
  connecting_services = []
}

service "http" "logger" {
  address            = "logger.svc.local"
  local_port         = 5001
  allow_all_services = true
}

service "http" "db" {
  address    = "db.svc.external"
  local_port = 5002
  connecting_services = [
    "api"
  ]
}

ingress "api" {
  domains = [
    "api.example.com",
    "other.example.com",
  ]
}
