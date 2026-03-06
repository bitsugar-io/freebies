output "cluster_id" {
  description = "DOKS cluster ID"
  value       = digitalocean_kubernetes_cluster.freebie.id
}

output "cluster_endpoint" {
  description = "Kubernetes API server endpoint"
  value       = digitalocean_kubernetes_cluster.freebie.endpoint
}

output "kubeconfig" {
  description = "Kubeconfig for the cluster"
  value       = digitalocean_kubernetes_cluster.freebie.kube_config[0].raw_config
  sensitive   = true
}

output "registry_endpoint" {
  description = "Container registry endpoint"
  value       = digitalocean_container_registry.freebie.endpoint
}
