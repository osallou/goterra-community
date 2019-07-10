# Configure the OpenStack Provider
provider "openstack" {
  user_name   = var.user_name
  password    = var.password

  tenant_name = var.tenant_name
  tenant_id = var.tenant_id
  auth_url    = var.auth_url
  domain_id = var.domain_id
  project_domain_id = var.project_domain_id
  user_domain_id = var.user_domain_id

}

# Configure the Goterra Provider
#provider "goterra" {
#  address = var.goterra_url
#  apikey = var.goterra_apikey
#}

#resource "goterra_deployment" "my-deploy" {
#}

#resource "goterra_application" "share" {
#  name = "share"
#  recipes = var.recipes_master
#  recipe_tags = []
#  deployment = "${goterra_deployment.my-deploy.id}"
#  deployment_token = "${goterra_deployment.my-deploy.token}"
#  application = var.goterra_application
#  namespace = var.goterra_namespace
#
#  depends_on = ["goterra_deployment.my-deploy"]
#}

resource "openstack_sharedfilesystem_share_v2" "share_1" {
  name             = "${var.volume_name}_${var.namespace}"
  description      = "share filesystem"
  share_proto      = "NFS"
  size             = var.volume_size

  #depends_on = ["goterra_deployment.my-deploy"]
}


output "myshare_id" {
  value = "${openstack_sharedfilesystem_share_v2.share_1.id}"
}

output "myshare_path" {
  value = "${openstack_sharedfilesystem_share_v2.share_1.export_locations[0].path}"
}
