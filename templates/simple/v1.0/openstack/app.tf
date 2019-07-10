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
provider "goterra" {
  address = var.goterra_url
  apikey = var.goterra_apikey
}

resource "goterra_deployment" "my-deploy" {
}

resource "goterra_application" "master" {
  name = "master"
  recipes = var.recipes_vm
  recipe_tags = []
  deployment = "${goterra_deployment.my-deploy.id}"
  deployment_token = "${goterra_deployment.my-deploy.token}"
  application = var.goterra_application
  namespace = var.goterra_namespace

  depends_on = ["goterra_deployment.my-deploy"]

}

resource "openstack_compute_instance_v2" "master" {
  name = "vm_${var.goterra_namespace}"
  image_id = var.image_id
  flavor_name = var.flavor_name
  key_pair = var.key_pair
  security_groups = ["default"]
  network {
    name = var.network
  }

  block_device {
    uuid                  = var.image_id
    source_type           = "image"
    destination_type      = "local"
    boot_index            = 0
    delete_on_termination = true
  }

  block_device {
    source_type           = "blank"
    destination_type      = "volume"
    volume_size           = var.volume_size
    boot_index            = 1
    delete_on_termination = true
  }

  user_data = file("${path.module}/${goterra_application.master.cloudinit}")
}

resource "openstack_networking_floatingip_v2" "masterfloatip" {
  pool = var.public_ip_pool
  count = var.feature_ip_public == "1" ? 1 : 0
}

resource "openstack_compute_floatingip_associate_v2" "fip_1" {
  count = var.feature_ip_public == "1" ? 1 : 0
  floating_ip = "${openstack_networking_floatingip_v2.masterfloatip[0].address}"
  instance_id = "${openstack_compute_instance_v2.master.id}"
  depends_on = ["openstack_compute_instance_v2.master"]

}

resource "goterra_push" "masterip" {
  address = "${goterra_deployment.my-deploy.address}"
  token = "${goterra_deployment.my-deploy.token}"
  deployment = "${goterra_deployment.my-deploy.id}"
  key = "masterip"
  value = "${openstack_compute_instance_v2.master.network.0.fixed_ip_v4}"

  depends_on = ["openstack_compute_instance_v2.master"]
}


output "masterip" {
  value = "${openstack_compute_instance_v2.master.network.0.fixed_ip_v4}"
}

output "masterpublicip" {
  value = "${openstack_networking_floatingip_v2.masterfloatip[0].address}"
}

output "deployment_id" {
  value = "${goterra_deployment.my-deploy.id}"
  depends_on = ["goterra_deployment.my-deploy"]

}
