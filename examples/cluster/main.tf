terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
    }
  }
}

provider libvirt {
  uri = "qemu:///system"
}

resource libvirt_pool khutulun {
  name = "khutulun"
  type = "dir"
  path = "/tmp/libvirt-pool-khutulun"
}

resource libvirt_network khutulun {
  name = "khutulun"
  mode = "nat"
  addresses = [ "192.168.123.0/24" ]
  autostart = true
}

data template_file cloud_init {
  template = "${file("${path.module}/cloud_init.cfg")}"
}

resource libvirt_volume centos-qcow2 {
  name = "centos.qcow2"
  pool = libvirt_pool.khutulun.name
  source = "https://cloud.centos.org/centos/8-stream/x86_64/images/CentOS-Stream-GenericCloud-8-20220125.1.x86_64.qcow2"
  format = "qcow2"
}

resource libvirt_cloudinit_disk cloud_init {
  name = "cloud_init.iso"
  pool = libvirt_pool.khutulun.name
  user_data = data.template_file.cloud_init.rendered
}

resource libvirt_domain host1 {
  name = "host1"
  
  memory = "2048"
  vcpu = 2
  autostart = true

  network_interface {
    network_name = libvirt_network.khutulun.name
  }

  disk {
    volume_id = libvirt_volume.centos-qcow2.id
  }

  cloudinit = libvirt_cloudinit_disk.cloud_init.id

  console {
    type = "pty"
    target_type = "serial"
    target_port = "0"
  }

  graphics {
    type = "spice"
    listen_type = "address"
    autoport = true
  }
}
