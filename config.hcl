global {
  log_directory      = "./logs"
  ingress_port       = 8080
  connection_timeout = "2s"
}

service "http" "api" {
  address             = "api-svc.default.svc"
  local_port          = 5000
  connecting_services = []
}

service "http" "logger" {
  address            = "logger-svc.default.svc"
  local_port         = 5001
  allow_all_services = true
}

egress "http" "db" {
  address    = "db.external"
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
