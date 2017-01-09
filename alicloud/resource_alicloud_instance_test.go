package alicloud

import (
	"fmt"
	"testing"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
)

func TestAccAlicloudInstance_basic(t *testing.T) {
	var instance ecs.InstanceAttributesType

	testCheck := func(*terraform.State) error {
		log.Printf("[WARN] instances: %#v", instance)
		if instance.ZoneId == "" {
			return fmt.Errorf("bad availability zone")
		}
		if len(instance.SecurityGroupIds.SecurityGroupId) == 0 {
			return fmt.Errorf("no security group: %#v", instance.SecurityGroupIds.SecurityGroupId)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		// module name
		IDRefreshName: "alicloud_instance.foo",

		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"alicloud_instance.foo", &instance),
					testCheck,
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"image_id",
						"ubuntu_140405_32_40G_cloudinit_20161115.vhd"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"instance_name",
						"test_foo"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"internet_charge_type",
						"PayByBandwidth"),
				),
			},

			// test for multi steps
			resource.TestStep{
				Config: testAccInstanceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"alicloud_instance.foo", &instance),
					testCheck,
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"image_id",
						"ubuntu_140405_32_40G_cloudinit_20161115.vhd"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"instance_name",
						"test_foo"),
				),
			},
		},
	})

}

func TestAccAlicloudInstance_vpc(t *testing.T) {
	var instance ecs.InstanceAttributesType

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: "alicloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"alicloud_instance.foo", &instance),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"system_disk_category",
						"cloud_efficiency"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"internet_charge_type",
						"PayByTraffic"),
				),
			},
		},
	})
}

func TestAccAlicloudInstance_multipleRegions(t *testing.T) {
	var instance ecs.InstanceAttributesType

	// multi provideris
	var providers []*schema.Provider
	providerFactories := map[string]terraform.ResourceProviderFactory{
		"alicloud": func() (terraform.ResourceProvider, error) {
			p := Provider()
			providers = append(providers, p.(*schema.Provider))
			return p, nil
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckInstanceDestroyWithProviders(&providers),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigMultipleRegions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExistsWithProviders(
						"alicloud_instance.foo", &instance, &providers),
					testAccCheckInstanceExistsWithProviders(
						"alicloud_instance.bar", &instance, &providers),
				),
			},
		},
	})
}

func TestAccAlicloudInstance_multiSecurityGroup(t *testing.T) {
	var instance ecs.InstanceAttributesType

	testCheck := func(sgCount int) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if len(instance.SecurityGroupIds.SecurityGroupId) < 0 {
				return fmt.Errorf("no security group: %#v", instance.SecurityGroupIds.SecurityGroupId)
			}

			if len(instance.SecurityGroupIds.SecurityGroupId) < sgCount {
				return fmt.Errorf("less security group: %#v", instance.SecurityGroupIds.SecurityGroupId)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		// module name
		IDRefreshName: "alicloud_instance.foo",

		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfig_multiSecurityGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"alicloud_instance.foo", &instance),
					testCheck(2),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"image_id",
						"ubuntu_140405_32_40G_cloudinit_20161115.vhd"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"instance_name",
						"test_foo"),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfig_multiSecurityGroup_add,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"alicloud_instance.foo", &instance),
					testCheck(3),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"image_id",
						"ubuntu_140405_32_40G_cloudinit_20161115.vhd"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"instance_name",
						"test_foo"),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfig_multiSecurityGroup_remove,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"alicloud_instance.foo", &instance),
					testCheck(1),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"image_id",
						"ubuntu_140405_32_40G_cloudinit_20161115.vhd"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"instance_name",
						"test_foo"),
				),
			},
		},
	})

}

func TestAccAlicloudInstance_multiSecurityGroupByCount(t *testing.T) {
	var instance ecs.InstanceAttributesType

	testCheck := func(sgCount int) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if len(instance.SecurityGroupIds.SecurityGroupId) < 0 {
				return fmt.Errorf("no security group: %#v", instance.SecurityGroupIds.SecurityGroupId)
			}

			if len(instance.SecurityGroupIds.SecurityGroupId) < sgCount {
				return fmt.Errorf("less security group: %#v", instance.SecurityGroupIds.SecurityGroupId)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		// module name
		IDRefreshName: "alicloud_instance.foo",

		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfig_multiSecurityGroupByCount,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"alicloud_instance.foo", &instance),
					testCheck(2),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"image_id",
						"ubuntu_140405_32_40G_cloudinit_20161115.vhd"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"instance_name",
						"test_foo"),
				),
			},
		},
	})

}

func TestAccAlicloudInstance_NetworkInstanceSecurityGroups(t *testing.T) {
	var instance ecs.InstanceAttributesType

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: "alicloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceNetworkInstanceSecurityGroups,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"alicloud_instance.foo", &instance),
				),
			},
		},
	})
}

func TestAccAlicloudInstance_tags(t *testing.T) {
	var instance ecs.InstanceAttributesType

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckInstanceConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("alicloud_instance.foo", &instance),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"tags.foo",
						"bar"),
				),
			},

			resource.TestStep{
				Config: testAccCheckInstanceConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("alicloud_instance.foo", &instance),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"tags.foo",
						""),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"tags.bar",
						"zzz"),
				),
			},
		},
	})
}

func TestAccAlicloudInstance_update(t *testing.T) {
	var instance ecs.InstanceAttributesType

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckInstanceConfigOrigin,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("alicloud_instance.foo", &instance),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"instance_name",
						"instance_foo"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"host_name",
						"host-foo"),
				),
			},

			resource.TestStep{
				Config: testAccCheckInstanceConfigOriginUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("alicloud_instance.foo", &instance),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"instance_name",
						"instance_bar"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"host_name",
						"host-bar"),
				),
			},
		},
	})
}

func TestAccAlicloudInstance_privateIP(t *testing.T) {
	var instance ecs.InstanceAttributesType

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			privateIP := instance.VpcAttributes.PrivateIpAddress.IpAddress[0]
			if privateIP != "172.16.0.229" {
				return fmt.Errorf("bad private IP: %s", privateIP)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: "alicloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigPrivateIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("alicloud_instance.foo", &instance),
					testCheckPrivateIP(),
				),
			},
		},
	})
}

func TestAccAlicloudInstance_associatePublicIPAndPrivateIP(t *testing.T) {
	var instance ecs.InstanceAttributesType

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			privateIP := instance.VpcAttributes.PrivateIpAddress.IpAddress[0]
			if privateIP != "172.16.0.229" {
				return fmt.Errorf("bad private IP: %s", privateIP)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: "alicloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigAssociatePublicIPAndPrivateIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("alicloud_instance.foo", &instance),
					testCheckPrivateIP(),
				),
			},
		},
	})
}

func TestAccAlicloudInstance_vpcRule(t *testing.T) {
	var instance ecs.InstanceAttributesType

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: "alicloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVpcInstanceWithSecurityRule,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("alicloud_instance.foo", &instance),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"internet_charge_type",
						"PayByBandwidth"),
					resource.TestCheckResourceAttr(
						"alicloud_instance.foo",
						"internet_max_bandwidth_out",
						"5"),
				),
			},
		},
	})
}

func testAccCheckInstanceExists(n string, i *ecs.InstanceAttributesType) resource.TestCheckFunc {
	providers := []*schema.Provider{testAccProvider}
	return testAccCheckInstanceExistsWithProviders(n, i, &providers)
}

func testAccCheckInstanceExistsWithProviders(n string, i *ecs.InstanceAttributesType, providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		for _, provider := range *providers {
			// Ignore if Meta is empty, this can happen for validation providers
			if provider.Meta() == nil {
				continue
			}

			client := provider.Meta().(*AliyunClient)
			instance, err := client.QueryInstancesById(rs.Primary.ID)
			log.Printf("[WARN]get ecs instance %#v", instance)
			if err == nil && instance != nil {
				*i = *instance
				return nil
			}

			// Verify the error is what we want
			e, _ := err.(*common.Error)
			if e.ErrorResponse.Message == InstanceNotfound {
				continue
			}
			if err != nil {
				return err

			}
		}

		return fmt.Errorf("Instance not found")
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	return testAccCheckInstanceDestroyWithProvider(s, testAccProvider)
}

func testAccCheckInstanceDestroyWithProviders(providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, provider := range *providers {
			if provider.Meta() == nil {
				continue
			}
			if err := testAccCheckInstanceDestroyWithProvider(s, provider); err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckInstanceDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	client := provider.Meta().(*AliyunClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "alicloud_instance" {
			continue
		}

		// Try to find the resource
		instance, err := client.QueryInstancesById(rs.Primary.ID)
		if err == nil {
			if instance.Status != "" && instance.Status != "Stopped" {
				return fmt.Errorf("Found unstopped instance: %s", instance.InstanceId)
			}
		}

		// Verify the error is what we want
		e, _ := err.(*common.Error)
		if e.ErrorResponse.Message == InstanceNotfound {
			continue
		}

		return err
	}

	return nil
}

const testAccInstanceConfig = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
}

resource "alicloud_security_group" "tf_test_bar" {
	name = "tf_test_bar"
	description = "bar"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	system_disk_category = "cloud_ssd"

	instance_type = "ecs.n1.small"
	internet_charge_type = "PayByBandwidth"
	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]
	instance_name = "test_foo"

	tags {
		foo = "bar"
		work = "test"
	}
}
`
const testAccInstanceConfigVPC = `
resource "alicloud_vpc" "foo" {
  name = "tf_test_foo"
  cidr_block = "172.16.0.0/12"
}

resource "alicloud_vswitch" "foo" {
  vpc_id = "${alicloud_vpc.foo.id}"
  cidr_block = "172.16.0.0/21"
  availability_zone = "cn-beijing-b"
}

resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
	vpc_id = "${alicloud_vpc.foo.id}"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	vswitch_id = "${alicloud_vswitch.foo.id}"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	# series II
	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	system_disk_category = "cloud_efficiency"

	internet_charge_type = "PayByTraffic"
	internet_max_bandwidth_out = 5
	allocate_public_ip = true
	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]
	instance_name = "test_foo"
}

`
const testAccInstanceConfigMultipleRegions = `
provider "alicloud" {
	alias = "beijing"
	region = "cn-beijing"
}

provider "alicloud" {
	alias = "shanghai"
	region = "cn-shanghai"
}

resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	provider = "alicloud.beijing"
	description = "foo"
}

resource "alicloud_security_group" "tf_test_bar" {
	name = "tf_test_bar"
	provider = "alicloud.shanghai"
	description = "bar"
}

resource "alicloud_instance" "foo" {
  # cn-beijing
  provider = "alicloud.beijing"
  availability_zone = "cn-beijing-b"
  image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

  internet_charge_type = "PayByBandwidth"

  instance_type = "ecs.n1.medium"
  io_optimized = "optimized"
  system_disk_category = "cloud_efficiency"
  security_groups = ["${alicloud_security_group.tf_test_foo.id}"]
  instance_name = "test_foo"
}

resource "alicloud_instance" "bar" {
	# cn-shanghai
	provider = "alicloud.shanghai"
	availability_zone = "cn-shanghai-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	internet_charge_type = "PayByBandwidth"

	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	system_disk_category = "cloud_efficiency"
	security_groups = ["${alicloud_security_group.tf_test_bar.id}"]
	instance_name = "test_bar"
}
`

const testAccInstanceConfig_multiSecurityGroup = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
}

resource "alicloud_security_group" "tf_test_bar" {
	name = "tf_test_bar"
	description = "bar"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	instance_type = "ecs.s2.large"
	internet_charge_type = "PayByBandwidth"
	security_groups = ["${alicloud_security_group.tf_test_foo.id}", "${alicloud_security_group.tf_test_bar.id}"]
	instance_name = "test_foo"
}`

const testAccInstanceConfig_multiSecurityGroup_add = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
}

resource "alicloud_security_group" "tf_test_bar" {
	name = "tf_test_bar"
	description = "bar"
}

resource "alicloud_security_group" "tf_test_add_sg" {
	name = "tf_test_add_sg"
	description = "sg"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	instance_type = "ecs.s2.large"
	internet_charge_type = "PayByBandwidth"
	security_groups = ["${alicloud_security_group.tf_test_foo.id}", "${alicloud_security_group.tf_test_bar.id}",
				"${alicloud_security_group.tf_test_add_sg.id}"]
	instance_name = "test_foo"
}
`

const testAccInstanceConfig_multiSecurityGroup_remove = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	instance_type = "ecs.s2.large"
	internet_charge_type = "PayByBandwidth"
	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]
	instance_name = "test_foo"
}
`

const testAccInstanceConfig_multiSecurityGroupByCount = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	count = 2
	description = "foo"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	instance_type = "ecs.s2.large"
	internet_charge_type = "PayByBandwidth"
	security_groups = ["${alicloud_security_group.tf_test_foo.*.id}"]
	instance_name = "test_foo"
}
`

const testAccInstanceNetworkInstanceSecurityGroups = `
resource "alicloud_vpc" "foo" {
  name = "tf_test_foo"
  cidr_block = "172.16.0.0/12"
}

resource "alicloud_vswitch" "foo" {
  vpc_id = "${alicloud_vpc.foo.id}"
  cidr_block = "172.16.0.0/21"
  availability_zone = "cn-beijing-b"
}

resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
	vpc_id = "${alicloud_vpc.foo.id}"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	vswitch_id = "${alicloud_vswitch.foo.id}"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	# series II
	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	system_disk_category = "cloud_efficiency"

	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]
	instance_name = "test_foo"

	internet_max_bandwidth_out = 5
	allocate_public_ip = "true"
	internet_charge_type = "PayByBandwidth"
}
`
const testAccCheckInstanceConfigTags = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	# series II
	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	internet_charge_type = "PayByBandwidth"
	system_disk_category = "cloud_efficiency"

	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]
	instance_name = "test_foo"

	tags {
		foo = "bar"
	}
}
`

const testAccCheckInstanceConfigTagsUpdate = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	# series II
	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	internet_charge_type = "PayByBandwidth"
	system_disk_category = "cloud_efficiency"

	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]
	instance_name = "test_foo"

	tags {
		bar = "zzz"
	}
}
`
const testAccCheckInstanceConfigOrigin = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	# series II
	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	internet_charge_type = "PayByBandwidth"
	system_disk_category = "cloud_efficiency"

	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]

	instance_name = "instance_foo"
	host_name = "host-foo"
}
`

const testAccCheckInstanceConfigOriginUpdate = `
resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"

	# series II
	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	internet_charge_type = "PayByBandwidth"
	system_disk_category = "cloud_efficiency"

	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]
	
	instance_name = "instance_bar"
	host_name = "host-bar"
}
`

const testAccInstanceConfigPrivateIP = `
resource "alicloud_vpc" "foo" {
  name = "tf_test_foo"
  cidr_block = "172.16.0.0/12"
}

resource "alicloud_vswitch" "foo" {
  vpc_id = "${alicloud_vpc.foo.id}"
  cidr_block = "172.16.0.0/24"
  availability_zone = "cn-beijing-b"
}

resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
	vpc_id = "${alicloud_vpc.foo.id}"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]

	vswitch_id = "${alicloud_vswitch.foo.id}"
	private_ip = "172.16.0.229"

	# series II
	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	system_disk_category = "cloud_efficiency"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"
	instance_name = "test_foo"
}
`
const testAccInstanceConfigAssociatePublicIPAndPrivateIP = `
resource "alicloud_vpc" "foo" {
  name = "tf_test_foo"
  cidr_block = "172.16.0.0/12"
}

resource "alicloud_vswitch" "foo" {
  vpc_id = "${alicloud_vpc.foo.id}"
  cidr_block = "172.16.0.0/24"
  availability_zone = "cn-beijing-b"
}

resource "alicloud_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"
	vpc_id = "${alicloud_vpc.foo.id}"
}

resource "alicloud_instance" "foo" {
	# cn-beijing
	availability_zone = "cn-beijing-b"
	security_groups = ["${alicloud_security_group.tf_test_foo.id}"]

	vswitch_id = "${alicloud_vswitch.foo.id}"
	private_ip = "172.16.0.229"
	allocate_public_ip = "true"
	internet_max_bandwidth_out = 5
	internet_charge_type = "PayByBandwidth"

	# series II
	instance_type = "ecs.n1.medium"
	io_optimized = "optimized"
	system_disk_category = "cloud_efficiency"
	image_id = "ubuntu_140405_32_40G_cloudinit_20161115.vhd"
	instance_name = "test_foo"
}
`
const testAccVpcInstanceWithSecurityRule = `
resource "alicloud_vpc" "foo" {
  name = "tf_test_foo"
  cidr_block = "10.1.0.0/21"
}

resource "alicloud_vswitch" "foo" {
  vpc_id = "${alicloud_vpc.foo.id}"
  cidr_block = "10.1.1.0/24"
  availability_zone = "cn-beijing-c"
}

resource "alicloud_security_group" "tf_test_foo" {
    name = "tf_test_foo"
    description = "foo"
    vpc_id = "${alicloud_vpc.foo.id}"
}

resource "alicloud_security_group_rule" "ingress" {
  type = "ingress"
  ip_protocol = "tcp"
  nic_type = "intranet"
  policy = "accept"
  port_range = "22/22"
  priority = 1
  security_group_id = "${alicloud_security_group.tf_test_foo.id}"
  cidr_ip = "0.0.0.0/0"
}

resource "alicloud_instance" "foo" {
    # cn-beijing
    availability_zone = "cn-beijing-c"
    security_groups = ["${alicloud_security_group.tf_test_foo.id}"]

    vswitch_id = "${alicloud_vswitch.foo.id}"
    allocate_public_ip = true

    # series II
    instance_charge_type = "PostPaid"
    instance_type = "ecs.n1.small"
    internet_charge_type = "PayByBandwidth"
    internet_max_bandwidth_out = 5

    system_disk_category = "cloud_efficiency"
    image_id = "ubuntu_140405_64_40G_cloudinit_20161115.vhd"
    instance_name = "test_foo"
}

`
