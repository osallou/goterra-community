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
  recipe_tags = ["master", "storage"]
  deployment = "${goterra_deployment.my-deploy.id}"
  deployment_token = "${goterra_deployment.my-deploy.token}"
  application = var.goterra_application
  namespace = var.goterra_namespace

  depends_on = ["goterra_deployment.my-deploy"]

}

resource "goterra_application" "slave" {
  name = "slave"
  recipe_tags = ["slave"]
  deployment = "${goterra_deployment.my-deploy.id}"
  deployment_token = "${goterra_deployment.my-deploy.token}"
  application = var.goterra_application
  namespace = var.goterra_namespace

  depends_on = ["goterra_deployment.my-deploy"]

}

resource "openstack_compute_instance_v2" "k3smaster" {
  name = "k3smaster"
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


data "goterra_deployment" "k3stoken" {
    timeout = 900
    deployment = "${goterra_deployment.my-deploy.id}"
    token = "${goterra_deployment.my-deploy.token}"
    key = "k3stoken"

    depends_on = ["openstack_compute_instance_v2.k3smaster"]
}

resource "goterra_push" "masterip" {
  address = "${goterra_deployment.my-deploy.address}"
  token = "${goterra_deployment.my-deploy.token}"
  deployment = "${goterra_deployment.my-deploy.id}"
  key = "masterip"
  value = "${openstack_compute_instance_v2.k3smaster.network.0.fixed_ip_v4}"

  depends_on = ["openstack_compute_instance_v2.k3smaster"]
}

resource "openstack_compute_instance_v2" "k3sslave" {

  name = "k3sslave${count.index}"
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

output "k3stoken" {
  value = "${data.goterra_deployment.k3stoken.data}"
  depends_on = ["data.goterra_deployment.k3stoken"]
}

output "masterip" {
  value = "${openstack_compute_instance_v2.k3smaster.network.0.fixed_ip_v4}"
}

output "slavesip" {
  value = ["${openstack_compute_instance_v2.k3sslave.*.network.0.fixed_ip_v4}"]
}

output "deployment_id" {
  value = "${goterra_deployment.my-deploy.id}"
  depends_on = ["goterra_deployment.my-deploy"]

}
