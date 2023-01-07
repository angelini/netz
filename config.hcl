global {
  log_directory      = "./logs"
  external_port      = 8080
  connection_timeout = "2s"
}

service "http" "api" {
  address             = "www.envoyproxy.io:80"
  local_port          = 4000
  connecting_services = []
}

service "http" "logger" {
  address            = "www.example.com:80"
  local_port         = 4001
  allow_all_services = true
}

service "http" "db" {
  address    = "www.example.com:80"
  local_port = 4002
  connecting_services = [
    "api"
  ]
}

