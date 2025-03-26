resource "cloudfoundry_network_policy" "policy" {
  policies = [
    {
      source_app      = "16b53647-9709-44bf-91b2-116de83ffd3d"
      destination_app = "41048361-adc7-4686-9115-36b16d8df12c"
      port            = "61443"
      protocol        = "tcp"
    },
    {
      source_app      = "16b53647-9709-44bf-91b2-116de83ffd3d"
      destination_app = "41048361-adc7-4686-9115-36b16d8df12c"
      port            = "8090-8092"
      protocol        = "udp"
    }
  ]
}
