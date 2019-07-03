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
  recipes = var.recipes_master
  recipe_tags = []
  deployment = "${goterra_deployment.my-deploy.id}"
  deployment_token = "${goterra_deployment.my-deploy.token}"
  application = var.goterra_application
  namespace = var.goterra_namespace

  depends_on = ["goterra_deployment.my-deploy"]

}

resource "goterra_application" "slave" {
  name = "slave"
  recipes = var.recipes_slave
  recipe_tags = []
  deployment = "${goterra_deployment.my-deploy.id}"
  deployment_token = "${goterra_deployment.my-deploy.token}"
  application = var.goterra_application
  namespace = var.goterra_namespace

  depends_on = ["goterra_deployment.my-deploy"]

}

resource "openstack_compute_instance_v2" "master" {
  name = "master"
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
    volume_size           = 1
    boot_index            = 1
    delete_on_termination = true
  }

  user_data = file("${path.module}/${goterra_application.master.cloudinit}")
}


resource "goterra_push" "masterip" {
  address = "${goterra_deployment.my-deploy.address}"
  token = "${goterra_deployment.my-deploy.token}"
  deployment = "${goterra_deployment.my-deploy.id}"
  key = "masterip"
  value = "${openstack_compute_instance_v2.master.network.0.fixed_ip_v4}"

  depends_on = ["openstack_compute_instance_v2.master"]
}

resource "openstack_compute_instance_v2" "slave" {

  name = "slave${count.index}"
  image_id = var.image_id
  flavor_name = var.flavor_name
  key_pair = var.key_pair
  security_groups = ["default"]

  user_data = file("${path.module}/${goterra_application.slave.cloudinit}")

  count = var.slave_count

  depends_on = ["goterra_push.masterip"]

  network {
    name = var.network
  }
}

output "masterip" {
  value = "${openstack_compute_instance_v2.master.network.0.fixed_ip_v4}"
}

output "slavesip" {
  value = ["${openstack_compute_instance_v2.slave.*.network.0.fixed_ip_v4}"]
}

output "deployment_id" {
  value = "${goterra_deployment.my-deploy.id}"
  depends_on = ["goterra_deployment.my-deploy"]

}
