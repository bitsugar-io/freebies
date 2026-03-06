resource "digitalocean_container_registry" "freebie" {
  name                   = "freebie"
  subscription_tier_slug = "starter"
  region                 = var.region
}
