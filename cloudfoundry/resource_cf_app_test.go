package cloudfoundry

import (
	"fmt"
	"regexp"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const appResourceSpringMusicTemplate = `

data "cf_domain" "local" {
    name = "%s"
}
data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}
data "cf_service" "mysql" {
    name = "p-mysql"
}
data "cf_service" "rmq" {
    name = "p-rabbitmq"
}

resource "cf_route" "spring-music" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "spring-music"
}
resource "cf_service_instance" "db" {
	name = "db"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.mysql.service_plans.512mb}"
}
resource "cf_service_instance" "fs1" {
	name = "fs1"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.rmq.service_plans.standard}"
}
%%s
resource "cf_app" "spring-music" {
	name = "spring-music"
	space = "${data.cf_space.space.id}"
	memory = "768"
	disk_quota = "512"
	timeout = 1800

	url = "https://github.com/mevansam/spring-music/releases/download/v1.0/spring-music.war"

%%s
}
`

const appResourceSpringMusic = `

data "cf_domain" "local" {
    name = "%s"
}
data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}
data "cf_service" "mysql" {
    name = "p-mysql"
}
data "cf_service" "rmq" {
    name = "p-rabbitmq"
}

resource "cf_route" "spring-music" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "spring-music"
}
resource "cf_service_instance" "db" {
	name = "db"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.mysql.service_plans.512mb}"
}
resource "cf_service_instance" "fs1" {
	name = "fs1"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.rmq.service_plans.standard}"
}
resource "cf_app" "spring-music" {
	name = "spring-music"
	space = "${data.cf_space.space.id}"
	instances = "2"
	memory = "768"
	disk_quota = "512"
	timeout = 1800

	url = "https://github.com/mevansam/spring-music/releases/download/v1.0/spring-music.war"

	service_binding {
		service_instance = "${cf_service_instance.db.id}"
	}
	service_binding {
		service_instance = "${cf_service_instance.fs1.id}"
	}

	route {
		default_route = "${cf_route.spring-music.id}"
	}

	environment {
		TEST_VAR_1 = "testval1"
		TEST_VAR_2 = "testval2"
	}
}
`

const appResourceSpringMusicUpdate = `

data "cf_domain" "local" {
    name = "%s"
}
data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}
data "cf_service" "mysql" {
    name = "p-mysql"
}
data "cf_service" "rmq" {
    name = "p-rabbitmq"
}

resource "cf_route" "spring-music" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "spring-music"
}
resource "cf_service_instance" "db" {
	name = "db"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.mysql.service_plans.512mb}"
}
resource "cf_service_instance" "fs1" {
	name = "fs1"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.rmq.service_plans.standard}"
}
resource "cf_service_instance" "fs2" {
	name = "fs2"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.rmq.service_plans.standard}"
}
resource "cf_app" "spring-music" {
	name = "spring-music-updated"
	space = "${data.cf_space.space.id}"
	instances = "1"
	memory = "1024"
	disk_quota = "1024"
	timeout = 1800

	url = "https://github.com/mevansam/spring-music/releases/download/v1.0/spring-music.war"

	service_binding {
		service_instance = "${cf_service_instance.db.id}"
	}
	service_binding {
		service_instance = "${cf_service_instance.fs2.id}"
	}
	service_binding {
		service_instance = "${cf_service_instance.fs1.id}"
	}

	route {
		default_route = "${cf_route.spring-music.id}"
	}

	environment {
		TEST_VAR_1 = "testval1"
		TEST_VAR_2 = "testval2"
	}
}
`

const appResourceSpringMusicBlueGreenUpdate = `

data "cf_domain" "local" {
    name = "%s"
}
data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}
data "cf_service" "mysql" {
    name = "p-mysql"
}
data "cf_service" "rmq" {
    name = "p-rabbitmq"
}

resource "cf_route" "spring-music" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "spring-music"
}
resource "cf_route" "spring-music-stage" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "spring-music-stage"
}
resource "cf_service_instance" "db" {
	name = "db"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.mysql.service_plans.512mb}"
}
resource "cf_service_instance" "fs1" {
	name = "fs1"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.rmq.service_plans.standard}"
}
resource "cf_service_instance" "fs2" {
	name = "fs2"
    space = "${data.cf_space.space.id}"
    service_plan = "${data.cf_service.rmq.service_plans.standard}"
}
resource "cf_app" "spring-music" {
	name = "spring-music-updated"
	space = "${data.cf_space.space.id}"
	instances ="3"
	memory = "1024"
	disk_quota = "1024"
	timeout = 1800

	url = "https://github.com/mevansam/spring-music/releases/download/v1.0/spring-music.war"

	service_binding {
		service_instance = "${cf_service_instance.db.id}"
	}
	service_binding {
		service_instance = "${cf_service_instance.fs2.id}"
	}
	service_binding {
		service_instance = "${cf_service_instance.fs1.id}"
	}

	route {
		live_route = "${cf_route.spring-music.id}"
		stage_route = "${cf_route.spring-music-stage.id}"
	}

	environment {
		TEST_VAR_1 = "testval1"
		TEST_VAR_2 = "testval2"
	}

	blue_green = {
		enable = true
	}
}
`

const appResourceWithMultiplePorts = `

data "cf_domain" "local" {
    name = "%s"
}
data "cf_org" "org" {
    name = "pcfdev-org"
}
data "cf_space" "space" {
    name = "pcfdev-space"
	org = "${data.cf_org.org.id}"
}

resource "cf_app" "test-app" {
	name = "test-app"
	space = "${data.cf_space.space.id}"
	timeout = 1800
	ports = [ 8888, 9999 ]
	buildpack = "binary_buildpack"
	command = "chmod 0755 test-app && ./test-app --ports=8888,9999"
	health_check_type = "process"

	github_release {
		owner = "mevansam"
		repo = "test-app"
		filename = "test-app"
		version = "v0.0.1"
	}
}
resource "cf_route" "test-app-8888" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "test-app-8888"

	target {
		app = "${cf_app.test-app.id}"
		port = 8888
	}
}
resource "cf_route" "test-app-9999" {
	domain = "${data.cf_domain.local.id}"
	space = "${data.cf_space.space.id}"
	hostname = "test-app-9999"

	target {
		app = "${cf_app.test-app.id}"
		port = 9999
	}
}
`

func TestAccApp_app1(t *testing.T) {

	refApp := "cf_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceSpringMusic, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "2"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "2"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_1", "testval1"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_2", "testval2"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckResourceAttr(refApp, "service_binding.#", "2"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(appResourceSpringMusicUpdate, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music-updated"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "1024"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "1024"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "2"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_1", "testval1"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_2", "testval2"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckResourceAttr(refApp, "service_binding.#", "3"),
					),
				},
			},
		})
}
func TestAccApp_app1_bluegreen(t *testing.T) {

	refApp := "cf_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceSpringMusic, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "2"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "2"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_1", "testval1"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_2", "testval2"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckResourceAttr(refApp, "service_binding.#", "2"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(appResourceSpringMusicBlueGreenUpdate, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music-updated"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "3"),
						resource.TestCheckResourceAttr(refApp, "memory", "1024"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "1024"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "2"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_1", "testval1"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_2", "testval2"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckResourceAttr(refApp, "service_binding.#", "3"),
					),
				},
			},
		})
}
func TestAccApp_app2(t *testing.T) {

	refApp := "cf_app.test-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"test-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceWithMultiplePorts, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {
							responses := []string{"8888"}
							if err = assertHTTPResponse("https://test-app-8888."+defaultAppDomain()+"/port", 200, &responses); err != nil {
								return err
							}
							responses = []string{"9999"}
							if err = assertHTTPResponse("https://test-app-9999."+defaultAppDomain()+"/port", 200, &responses); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "test-app"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "2"),
						resource.TestCheckResourceAttr(refApp, "ports.8888", "8888"),
						resource.TestCheckResourceAttr(refApp, "ports.9999", "9999"),
					),
				},
			},
		})
}

func TestApp_OldStyleRoutes_failLiveStage(t *testing.T) {

	resource.Test(t,
		resource.TestCase{
			IsUnitTest:   true,
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("\\[REMOVED\\] Support for the non-default route has been removed."),
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						``,
						`route {
							live_route = "${cf_route.spring-music.id}"
						}`,
					),
				},

				resource.TestStep{
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("\\[REMOVED\\] Support for the non-default route has been removed."),
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						``,
						`route {
							stage_route = "${cf_route.spring-music.id}"
						}`,
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_updateTo(t *testing.T) {

	refApp := "cf_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						``,
						`route {
							default_route = "${cf_route.spring-music.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "1"),
						resource.TestCheckNoResourceAttr(refApp, "routes.#"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						``,
						`routes {
							route = "${cf_route.spring-music.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "0"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_updateToAndmore(t *testing.T) {

	refApp := "cf_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						``,
						`route {
							default_route = "${cf_route.spring-music.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "1"),
						resource.TestCheckNoResourceAttr(refApp, "routes.#"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						`resource "cf_route" "spring-music-2" {
							domain = "${data.cf_domain.local.id}"
							space = "${data.cf_space.space.id}"
							hostname = "spring-music-2"
						}`,
						`routes {
							route = "${cf_route.spring-music.id}"
						}
						routes {
							route = "${cf_route.spring-music-2.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://spring-music-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "0"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "2"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						`resource "cf_route" "spring-music-2" {
							domain = "${data.cf_domain.local.id}"
							space = "${data.cf_space.space.id}"
							hostname = "spring-music-2"
						}`,
						`routes {
							route = "${cf_route.spring-music.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://spring-music-2."+defaultAppDomain(), 404, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "0"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						`resource "cf_route" "spring-music-2" {
							domain = "${data.cf_domain.local.id}"
							space = "${data.cf_space.space.id}"
							hostname = "spring-music-2"
						}
						resource "cf_route" "spring-music-3" {
							domain = "${data.cf_domain.local.id}"
							space = "${data.cf_space.space.id}"
							hostname = "spring-music-3"
						}`,
						`routes {
							route = "${cf_route.spring-music-2.id}"
						}
						routes {
							route = "${cf_route.spring-music-3.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 404, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://spring-music-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://spring-music-3."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "0"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "2"),
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_Create(t *testing.T) {

	refApp := "cf_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						``,
						`routes {
							route = "${cf_route.spring-music.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_Change(t *testing.T) {

	refApp := "cf_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						``,
						`routes {
							route = "${cf_route.spring-music.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						`resource "cf_route" "spring-music-2" {
							domain = "${data.cf_domain.local.id}"
							space = "${data.cf_space.space.id}"
							hostname = "spring-music-2"
						}`,
						`routes {
							route = "${cf_route.spring-music-2.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 404, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://spring-music-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_Add(t *testing.T) {

	refApp := "cf_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						``,
						`routes {
							route = "${cf_route.spring-music.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceSpringMusicTemplate, defaultAppDomain()),
						`resource "cf_route" "spring-music-2" {
							domain = "${data.cf_domain.local.id}"
							space = "${data.cf_space.space.id}"
							hostname = "spring-music-2"
						}`,
						`routes {
							route = "${cf_route.spring-music.id}"
						}
						routes {
							route = "${cf_route.spring-music-2.id}"
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://spring-music."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://spring-music-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "spring-music"),
						resource.TestCheckResourceAttr(refApp, "space", defaultPcfDevSpaceID()),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "2"),
					),
				},
			},
		})
}

func testAccCheckAppExists(resApp string, validate func() error) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resApp]
		if !ok {
			return fmt.Errorf("app '%s' not found in terraform state", resApp)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resApp, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var (
			app             cfapi.CCApp
			routeMappings   []map[string]interface{}
			serviceBindings []map[string]interface{}
		)

		am := session.AppManager()
		rm := session.RouteManager()

		if app, err = am.ReadApp(id); err != nil {
			return err
		}
		session.Log.DebugMessage(
			"retrieved app for resource '%s' with id '%s': %# v",
			resApp, id, app)

		if err = assertEquals(attributes, "name", app.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "space", app.SpaceGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "instances", app.Instances); err != nil {
			return err
		}
		if err = assertEquals(attributes, "memory", app.Memory); err != nil {
			return err
		}
		if err = assertEquals(attributes, "disk_quota", app.DiskQuota); err != nil {
			return err
		}
		if err = assertEquals(attributes, "stack", app.StackGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "buildpack", app.Buildpack); err != nil {
			return err
		}
		if err = assertEquals(attributes, "command", app.Command); err != nil {
			return err
		}
		if err = assertEquals(attributes, "enable_ssh", app.EnableSSH); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_http_endpoint", app.HealthCheckHTTPEndpoint); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_type", app.HealthCheckType); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_timeout", app.HealthCheckTimeout); err != nil {
			return err
		}
		if err = assertMapEquals("environment", attributes, *app.Environment); err != nil {
			return err
		}

		if serviceBindings, err = am.ReadServiceBindingsByApp(id); err != nil {
			return err
		}
		session.Log.DebugMessage(
			"retrieved service bindings for app with id '%s': %# v",
			id, serviceBindings)

		if err = assertListEquals(attributes, "service_binding", len(serviceBindings),
			func(values map[string]string, i int) (match bool) {

				var binding map[string]interface{}

				serviceInstanceID := values["service_instance"]
				binding = nil

				for _, b := range serviceBindings {
					if serviceInstanceID == b["service_instance"] {
						binding = b
						break
					}
				}

				if binding != nil && values["binding_id"] == binding["binding_id"] {
					if err2 := assertMapEquals("credentials", values, binding["credentials"].(map[string]interface{})); err2 != nil {
						session.Log.LogMessage(
							"Credentials for service instance %s do not match: %s",
							serviceInstanceID, err2.Error())
						return false
					}
					return true
				}
				return false

			}); err != nil {
			return err
		}

		if routeMappings, err = rm.ReadRouteMappingsByApp(id); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved routes for app with id '%s': %# v",
			id, routeMappings)

		if err = validateRouteMappings(attributes, routeMappings); err != nil {
			return
		}

		err = validate()
		return
	}
}

func validateRouteMappings(attributes map[string]string, routeMappings []map[string]interface{}) (err error) {

	var (
		routeID, mappingID string
		mapping            map[string]interface{}

		ok bool
	)

	if _, isOldStyle := attributes["route.0.default_route"]; isOldStyle {
		routeKey := "route.0.default_route"
		routeMappingKey := "route.0.default_route_mapping_id"

		if routeID, ok = attributes[routeKey]; ok && len(routeID) > 0 {
			if mappingID, ok = attributes[routeMappingKey]; !ok || len(mappingID) == 0 {
				return fmt.Errorf("default route '%s' does not have a corresponding mapping id in the state", routeID)
			}

			mapping = nil
			for _, r := range routeMappings {
				if mappingID == r["mapping_id"] {
					mapping = r
					break
				}
			}
			if mapping == nil {
				return fmt.Errorf("unable to find route mapping with id '%s' for route '%s'", mappingID, routeID)
			}
			if routeID != mapping["route"] {
				return fmt.Errorf("route mapping with id '%s' does not map to route '%s'", mappingID, routeID)
			}
		}
		return err
	} else if _, isNewStyle := attributes["routes.0.route"]; isNewStyle {

		for i := 0; true; i++ {
			if routeID, ok := attributes[fmt.Sprintf("routes.%d.route", i)]; !ok {
				break
			} else {
				if mappingID, ok := attributes[fmt.Sprintf("routes.%d.mapping_id", i)]; !ok {
					return fmt.Errorf("Route with no mapping ID recored (routes.%d.route=%s)", i, routeID)
				} else {
					for _, r := range routeMappings {
						if mappingID == r["mapping_id"] {
							mapping = r
							break
						}
					}
					if mapping == nil {
						return fmt.Errorf("unable to find route mapping with id '%s' for route '%s'", mappingID, routeID)
					}
					if routeID != mapping["route"] {
						return fmt.Errorf("route mapping with id '%s' does not map to route '%s'", mappingID, routeID)
					}
				}
			}
		}
	}
	return nil
}

func testAccCheckAppDestroyed(apps []string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		for _, a := range apps {
			if _, err := session.AppManager().FindApp(a); err != nil {
				switch err.(type) {
				case *errors.ModelNotFoundError:
					continue
				default:
					return err
				}
			}
			return fmt.Errorf("app with name '%s' still exists in cloud foundry", a)
		}
		return nil
	}
}
