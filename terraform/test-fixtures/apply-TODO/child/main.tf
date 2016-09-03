variable "var" {}

resource "test_instance" "a" {}

resource "test_instance" "b" {
  param1 = "${test_instance.a.id}"
  param2 = "${var.var}"

  lifecycle { create_before_destroy = true }
}

resource "test_instance" "c" {
  var = "${test_instance.b.id}"
}
