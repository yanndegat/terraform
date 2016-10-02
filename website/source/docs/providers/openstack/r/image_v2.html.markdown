---
layout: "openstack"
page_title: "OpenStack: openstack_image_v2"
sidebar_current: "docs-openstack-resource-image-v2"
description: |-
  Manages a V2 Image resource within OpenStack Glance.
---

# openstack\_image_v2

Manages a V2 Image resource within OpenStack Glance.

## Example Usage

```
resource "openstack_image_v2" "rancheros" {
  name   = "RancherOS"
  image_source_url = "https://releases.rancher.com/os/latest/rancheros-openstack.img"
  container_format = "bare"
  disk_format = "qcow2"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Required) The region in which to obtain the V2 Glance client.
    A Glance client is needed to create an Image that can be used with
    a compute instance. If omitted, the `OS_REGION_NAME` environment variable
    is used. Changing this creates a new Image.

* `id` - (Optional) The ID of the Glance Image. If omitted, Glance will 
    generate a new UUID.

* `name` - (Required) The name of the image.

* `local_file_path` - (Optional) This is the filepath of the raw image file
   that will be uploaded to Glance. Conflicts with `image_source_url`.

* `image_source_url` - (Optional) This is the url of the raw image that will
   be downloaded in the `image_cache_path` before being uploaded to Glance. 
   Glance is able to download image from internet but the `gophercloud` library
   does not yet provide a way to do so.
   Conflicts with `local_file_path`.

* `image_cache_path` - (Optional) This is the directory where the images will
   be downloaded. Images will be stored with a filename corresponding to 
   the url's md5 hash. Defaults to "$HOME/.terraform/image_cache"
   
* `visibility` - (Optional) The visibility of the image. Must be one of 
   "public", "private", "community", or "shared". 

* `tags` - (Optional) The tags of the image. It must be a list of strings.

* `container_format` - (Required) The container format. Must be one of
   "ami", "ari", "aki", "bare", "ovf".

* `disk_format` - (Required) The disk format. Must be one of
   "ami", "ari", "aki", "vhd", "vmdk", "raw", "qcow2", "vdi", "iso".
   
* `min_disk_gb` - (Optional) Amount of disk space (in GB) required to boot image.
   Defaults to 0.

* `min_ram_mb` - (Optional) Amount of ram (in MB) required to boot image.
   Defauts to 0.
   
* `protected` - (Optional) If true, image will not be deletable.
   Defaults to false.

* `properties` - (Optional) A map of strings defining arbitrary properties 
   to associate with image.


## Attributes Reference

The following attributes are exported:

* `region` - See Argument Reference above.
* `id` - See Argument Reference above.
* `name` - See Argument Reference above.
* `visibility` - See Argument Reference above.
* `protected` - See Argument Reference above.
* `tags` - See Argument Reference above.
* `properties` - See Argument Reference above.
* `container_format` - See Argument Reference above.
* `disk_format` - See Argument Reference above.
* `min_disk_gb` - See Argument Reference above.
* `min_ram_mb` - See Argument Reference above.
* `protected` - See Argument Reference above.
* `owner` - The id of the openstack user who owns the image.
* `status` - The status of the image. It can be "queued", "active"
   or "saving".
* `checksum` - The checksum of the data associated with the image.
* `size_bytes` - The size in bytes of the data associated with the image.
* `metadata` - The metadata associated with the image.
   Image metadata allow for meaningfully define the image properties
   and tags. See http://docs.openstack.org/developer/glance/metadefs-concepts.html.
* `created_date` - The date the image was created.
* `last_update` - The date the image was last updated.
* `file` - the trailing path after the glance 
   endpoint that represent the location of the image 
   or the path to retrieve it.
* `schema` - The path to the JSON-schema that represent 
   the image or image entity

## Import

Images can be imported using the `id`, e.g.

```
$ terraform import openstack_image_v2.rancheros 89c60255-9bd6-460c-822a-e2b959ede9d2
```
