data "digitalocean_kubernetes_versions" "this" {
  version_prefix = "1.32."
}

resource "digitalocean_kubernetes_cluster" "freebie" {
  name    = "freebie"
  region  = var.region
  version = data.digitalocean_kubernetes_versions.this.latest_version

  node_pool {
    name       = "default"
    size       = var.node_size
    node_count = var.node_count
  }
}
