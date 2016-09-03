data "test_data_source" "foo" {
    value = "foo"
}

resource "test_instance" "foo" {
    key = "${data.test_data_source.foo.value}"
}

module "child" {
    source = "./child"
    var    = "${test_instance.foo.id}"
}
