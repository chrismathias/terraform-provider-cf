package cloudfoundry

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/cf/terminal"

	// "github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/repo"
)

// DefaultAppTimeout - Timeout (in seconds) when pushing apps to CF
const DefaultAppTimeout = 60

func resourceApp() *schema.Resource {

	return &schema.Resource{

		Create: resourceAppCreate,
		Read:   resourceAppRead,
		Update: resourceAppUpdate,
		Delete: resourceAppDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAppImport,
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ports": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Set:      resourceIntegerSet,
			},
			"instances": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"disk_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"stack": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"buildpack": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"command": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enable_ssh": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  DefaultAppTimeout,
			},
			"stopped": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"url": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"git", "github_release"},
			},
			"git": &schema.Schema{
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"url", "github_release"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"branch": &schema.Schema{
							Type:          schema.TypeString,
							Optional:      true,
							Default:       "master",
							ConflictsWith: []string{"git.tag"},
						},
						"tag": &schema.Schema{
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"git.branch"},
						},
						"user": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"password": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"key": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"github_release": &schema.Schema{
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"url", "git"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"owner": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"repo": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"token": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"version": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"filename": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"add_content": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"destination": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"service_binding": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_instance": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"params": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
						},
						"credentials": &schema.Schema{
							Type:     schema.TypeMap,
							Computed: true,
						},
						"binding_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"route": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_route": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_route_mapping_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"stage_route": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"stage_route_mapping_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"live_route": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"live_route_mapping_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"validation_script": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Removed:  "Use blue_green.validation_script instead.",
						},
					},
				},
			},
			"environment": &schema.Schema{
				Type:      schema.TypeMap,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"health_check_http_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"health_check_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateAppHealthCheckType,
			},
			"health_check_timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"disable_blue_green_deployment": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Removed:  "See new blue_green section instead to enable blue/green type updates.",
			},
			"blue_green": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"validation_script": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"deposed": {
				// This is not flagged as computed so that Terraform will always flag deposed resources as a change and allow us to attempt to clean them up
				Type:         schema.TypeMap,
				Optional:     true,
				Description:  "Do not use this, this field is meant for internal use only. (It is not flagged as Computed for technical reasons.)",
				ValidateFunc: validateAppDeposedMapEmpty,
			},
		},
	}
}

// func serviceBindingHash(d interface{}) int {
// 	return hashcode.String(d.(map[string]interface{})["service_instance"].(string))
// }

func validateAppHealthCheckType(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if value != "port" && value != "process" && value != "http" && value != "none" {
		errs = append(errs, fmt.Errorf("%q must be one of 'port', 'process', 'http' or 'none'", k))
	}
	return ws, errs
}

func validateAppDeposedMapEmpty(v interface{}, k string) (ws []string, errs []error) {
	if len(v.(map[string]interface{})) != 0 {
		errs = append(errs, fmt.Errorf("%q must not be set by the user", k))
	}
	return ws, errs
}

type cfAppConfig struct {
	app             cfapi.CCApp
	routeConfig     map[string]interface{}
	serviceBindings []map[string]interface{}
}

func resourceAppCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var app cfapi.CCApp
	app.Name = d.Get("name").(string)
	app.SpaceGUID = d.Get("space").(string)
	if v, ok := d.GetOk("ports"); ok {
		p := []int{}
		for _, vv := range v.(*schema.Set).List() {
			p = append(p, vv.(int))
		}
		app.Ports = &p
	}
	if v, ok := d.GetOk("instances"); ok {
		vv := v.(int)
		app.Instances = &vv
	}
	if v, ok := d.GetOk("memory"); ok {
		vv := v.(int)
		app.Memory = &vv
	}
	if v, ok := d.GetOk("disk_quota"); ok {
		vv := v.(int)
		app.DiskQuota = &vv
	}
	if v, ok := d.GetOk("stack"); ok {
		vv := v.(string)
		app.StackGUID = &vv
	}
	if v, ok := d.GetOk("buildpack"); ok {
		vv := v.(string)
		app.Buildpack = &vv
	}
	if v, ok := d.GetOk("command"); ok {
		vv := v.(string)
		app.Command = &vv
	}
	if v, ok := d.GetOk("enable_ssh"); ok {
		vv := v.(bool)
		app.EnableSSH = &vv
	}
	if v, ok := d.GetOk("health_check_http_endpoint"); ok {
		vv := v.(string)
		app.HealthCheckHTTPEndpoint = &vv
	}
	if v, ok := d.GetOk("health_check_type"); ok {
		vv := v.(string)
		app.HealthCheckType = &vv
	}
	if v, ok := d.GetOk("health_check_timeout"); ok {
		vv := v.(int)
		app.HealthCheckTimeout = &vv
	}
	if v, ok := d.GetOk("environment"); ok {
		vv := v.(map[string]interface{})
		app.Environment = &vv
	}

	appConfig := cfAppConfig{
		app: app,
	}

	if err := resourceAppCreateCfApp(d, meta, &appConfig); err != nil {
		return err
	}

	d.SetId(appConfig.app.ID)
	setAppArguments(appConfig.app, d)
	d.Set("service_binding", appConfig.serviceBindings)
	d.Set("route", []map[string]interface{}{appConfig.routeConfig})

	return nil
}

func resourceAppCreateCfApp(d *schema.ResourceData, meta interface{}, appConfig *cfAppConfig) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()
	rm := session.RouteManager()

	app := appConfig.app
	var (
		v interface{}

		appPath string

		defaultRoute, stageRoute, liveRoute string

		serviceBindings    []map[string]interface{}
		hasServiceBindings bool

		routeConfig    map[string]interface{}
		hasRouteConfig bool
	)

	// Download application binary / source asynchronously
	prepare := make(chan error)
	go func() {
		appPath, err = prepareApp(app, d, session.Log)
		prepare <- err
	}()

	if v, hasRouteConfig = d.GetOk("route"); hasRouteConfig {

		routeConfig = v.([]interface{})[0].(map[string]interface{})

		// ensure that if default route exists, that it is unbound or only bound to the existing application
		if defaultRoute, err = validateRoute(routeConfig, "default_route", d.Id(), rm); err != nil {
			return err
		}
		// ensure that if stage route exists, that it is unbound
		if stageRoute, err = validateRoute(routeConfig, "stage_route", "", rm); err != nil {
			return err
		}
		// ensure that if live route exists, that it is unbound or only bound to the existing application
		if liveRoute, err = validateRoute(routeConfig, "live_route", d.Id(), rm); err != nil {
			return err
		}

		if len(stageRoute) > 0 && len(liveRoute) > 0 {
		} else if len(stageRoute) > 0 || len(liveRoute) > 0 {
			err = fmt.Errorf("both 'stage_route' and 'live_route' need to be provided to deploy the app using blue-green routing")
			return err
		}
	}

	// Create application
	if app, err = am.CreateApp(app); err != nil {
		return err
	}
	// Delete application if an error occurs
	defer func() error {
		e := &err
		if *e != nil {
			return am.DeleteApp(app.ID, true)
		}
		return nil
	}()

	var addContent []map[string]interface{}
	if v, ok := d.GetOk("add_content"); ok {
		addContent = getListOfStructs(v)
	}
	// Upload application binary / source
	// asynchronously once download has completed
	if err = <-prepare; err != nil {
		return err
	}
	upload := make(chan error)
	go func() {
		err = am.UploadApp(app, appPath, addContent)
		upload <- err
	}()

	// Bind services
	if v, hasServiceBindings = d.GetOk("service_binding"); hasServiceBindings {
		if serviceBindings, err = addServiceBindings(app.ID, getListOfStructs(v), am, session.Log); err != nil {
			return err
		}
	}

	if d.Id() != "" {
		// we're doing a blue/green deployment, so we need to bind the stage_route
		if len(stageRoute) > 0 {
			if mappingID, err := rm.CreateRouteMapping(stageRoute, app.ID, nil); err != nil {
				return err
			} else {
				routeConfig["stage_route_mapping_id"] = mappingID
			}
		} else {
			return fmt.Errorf("stage_route is not defined, blue/green deployment failed")
		}
	} else {
		// Bind default_route
		if len(defaultRoute) > 0 {
			if mappingID, err := rm.CreateRouteMapping(defaultRoute, app.ID, nil); err != nil {
				return err
			} else {
				routeConfig["default_route_mapping_id"] = mappingID
			}
		}
		// Bind live_route
		if len(liveRoute) > 0 {
			if mappingID, err := rm.CreateRouteMapping(liveRoute, app.ID, nil); err != nil {
				return err
			} else {
				routeConfig["live_route_mapping_id"] = mappingID
			}
		}
	}

	timeout := time.Second * time.Duration(d.Get("timeout").(int))
	stopped := d.Get("stopped").(bool)

	// Start application if not stopped
	// state once upload has completed
	if err = <-upload; err != nil {
		return err
	}
	if !stopped {
		if err = am.StartApp(app.ID, timeout); err != nil {
			return err
		}
	}

	if app, err = am.ReadApp(app.ID); err != nil {
		return err
	}
	appConfig.app = app
	session.Log.DebugMessage("Created app state: %# v", app)

	if hasServiceBindings {
		appConfig.serviceBindings = serviceBindings
		session.Log.DebugMessage("Created service bindings: %# v", d.Get("service_binding"))
	}
	if hasRouteConfig {
		appConfig.routeConfig = routeConfig
		session.Log.DebugMessage("Created routes: %# v", d.Get("route"))
	}

	return err
}

func resourceAppRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id := d.Id()
	am := session.AppManager()

	var app cfapi.CCApp
	if app, err = am.ReadApp(id); err != nil {
		if strings.Contains(err.Error(), "status code: 404") {
			d.MarkNewResource()
			err = nil
		}
	} else {
		setAppArguments(app, d)
	}

	// check if any old deposed resources still exist
	if v, ok := d.GetOk("deposed"); ok {
		deposedResources := v.(map[string]interface{})
		for r, _ := range deposedResources {
			if _, err := am.ReadApp(r); err != nil {
				if strings.Contains(err.Error(), "status code: 404") {
					delete(deposedResources, r)
				}
			} else {
				delete(deposedResources, r)
			}
		}
		if err := d.Set("deposed", deposedResources); err != nil {
			return err
		}
	}

	return err
}

func resourceAppUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	// preseve deposed resources until we clean them up
	existingDeposed, _ := d.GetChange("deposed")
	d.Set("deposed", existingDeposed)

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	// TODO: clean-up old deposed resources

	app := cfapi.CCApp{}
	d.Partial(true)

	update := false // for changes where no restart is required
	app.Name = *getChangedValueString("name", &update, d)
	app.SpaceGUID = *getChangedValueString("space", &update, d)
	app.Instances = getChangedValueInt("instances", &update, d)
	app.EnableSSH = getChangedValueBool("enable_ssh", &update, d)
	app.HealthCheckHTTPEndpoint = getChangedValueString("health_check_http_endpoint", &update, d)
	app.HealthCheckType = getChangedValueString("health_check_type", &update, d)
	app.HealthCheckTimeout = getChangedValueInt("health_check_timeout", &update, d)

	restart := false // for changes where just a restart is required
	app.Ports = getChangedValueIntList("ports", &restart, d)
	app.Memory = getChangedValueInt("memory", &restart, d)
	app.DiskQuota = getChangedValueInt("disk_quota", &restart, d)
	app.Command = getChangedValueString("command", &restart, d)

	restage := false // for changes where a full restage is required
	app.Buildpack = getChangedValueString("buildpack", &restage, d)
	app.StackGUID = getChangedValueString("stack", &restage, d)
	app.Environment = getChangedValueMap("environment", &restage, d)

	blueGreen := false
	if v, ok := d.GetOk("blue_green"); ok {
		blueGreenConfig := v.([]interface{})[0].(map[string]interface{})
		if blueGreenEnabled, ok := blueGreenConfig["enable"]; ok && blueGreenEnabled.(bool) {
			if restart || restage || d.HasChange("service_binding") ||
				d.HasChange("url") || d.HasChange("git") || d.HasChange("github_release") || d.HasChange("add_content") {
				blueGreen = true
			}
		}
	}

	if blueGreen {
		err = resourceAppBlueGreenUpdate(d, meta, app)
	} else {
		// fall back to a standard update to the existing app
		err = resourceAppStandardUpdate(d, meta, app, update, restart, restage)
	}

	if err == nil {
		d.Partial(false)
	}

	return err
}

func resourceAppBlueGreenUpdate(d *schema.ResourceData, meta interface{}, newApp cfapi.CCApp) error {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()
	rm := session.RouteManager()

	var venerableApp cfapi.CCApp
	if v, err := am.ReadApp(d.Id()); err != nil {
		return err
	} else {
		venerableApp = v
	}

	// Update origin app name
	if venerableAppRefeshed, err := am.UpdateApp(cfapi.CCApp{ID: d.Id(), Name: venerableApp.Name + "-venerable"}); err != nil {
		return err
	} else {
		venerableApp = venerableAppRefeshed
	}

	appConfig := cfAppConfig{
		app: newApp,
	}
	appConfig.app.Instances = func(i int) *int { return &i }(1) // start the staged app with only one instance (we'll scale it up later)
	if err := resourceAppCreateCfApp(d, meta, &appConfig); err != nil {
		return err
	}
	appConfig.app.Instances = newApp.Instances // restore final expected instances count

	// TODO: Execute blue-green validation

	// now that we've passed validation, we've passed the point of no return
	d.SetId(appConfig.app.ID)
	d.SetPartial("url")
	d.SetPartial("git")
	d.SetPartial("github_release")
	d.SetPartial("add_content")
	d.SetPartial("service_binding")
	setAppArguments(appConfig.app, d)

	// ensure we keep track of the old application to clean it up later if we fail
	deposedResources := d.Get("deposed").(map[string]interface{})
	deposedResources[venerableApp.ID] = "application"
	d.Set("deposed", deposedResources)

	// Now bind the other routes to the new application instance and scale it up
	// Bind default_route
	if defaultRoute, err := validateRoute(appConfig.routeConfig, "default_route", venerableApp.ID, rm); err != nil {
		return err
	} else if len(defaultRoute) > 0 {
		if mappingID, err := rm.CreateRouteMapping(defaultRoute, appConfig.app.ID, nil); err != nil {
			return err
		} else {
			appConfig.routeConfig["default_route_mapping_id"] = mappingID
		}
	}
	// Bind live_route
	if liveRoute, err := validateRoute(appConfig.routeConfig, "live_route", venerableApp.ID, rm); err != nil {
		return err
	} else if len(liveRoute) > 0 {
		if mappingID, err := rm.CreateRouteMapping(liveRoute, appConfig.app.ID, nil); err != nil {
			return err
		} else {
			appConfig.routeConfig["live_route_mapping_id"] = mappingID
		}
	}
	d.SetPartial("route")

	var timeoutDuration time.Duration
	if v, ok := d.GetOk("timeout"); ok {
		vv := v.(int)
		timeoutDuration = time.Second * time.Duration(vv)
	}

	// now scale up the new app and scale down the old app
	venerableAppScale := cfapi.CCApp{
		ID:        venerableApp.ID,
		Name:      venerableApp.Name,
		Instances: venerableApp.Instances,
	}
	newAppScale := cfapi.CCApp{
		ID:        appConfig.app.ID,
		Name:      appConfig.app.Name,
		Instances: func(i int) *int { return &i }(1),
	}
	session.Log.DebugMessage("newApp.Instances: %d", *newApp.Instances)
	session.Log.DebugMessage("venerableApp.Instances: %d", *venerableAppScale.Instances)
	for *newAppScale.Instances < *newApp.Instances || *venerableAppScale.Instances > 1 {
		if *newAppScale.Instances < *newApp.Instances {
			// scale up new
			*newAppScale.Instances++
			session.Log.DebugMessage("Scaling up new app %s to instance count %d", newAppScale.ID, *newAppScale.Instances)
			if _, err := am.UpdateApp(newAppScale); err != nil {
				return err
			}
			if *(appConfig.app.State) != "STOPPED" {
				time.Sleep(time.Second * time.Duration(15))
				// TODO: fix this wait
				am.WaitForAppToStart(newAppScale, timeoutDuration)
			}
		}

		if *venerableAppScale.Instances > 1 {
			// scale down old
			*venerableAppScale.Instances--
			session.Log.DebugMessage("Scaling down venerable app %s to instance count %d", venerableAppScale.ID, *venerableAppScale.Instances)
			if _, err := am.UpdateApp(venerableAppScale); err != nil {
				return err
			}
			if *venerableApp.State != "STOPPED" {
				time.Sleep(time.Second * time.Duration(5))
				// TODO: wait for instance to stop
			}
		}
	}

	// now delete the old application
	if err := am.DeleteApp(venerableAppScale.ID, true); err != nil {
		return err
	} else {
		deposedResources := d.Get("deposed").(map[string]interface{})
		delete(deposedResources, venerableApp.ID)
		d.Set("deposed", deposedResources)
	}

	// TODO: unmap stage route?

	return nil
}

func resourceAppStandardUpdate(d *schema.ResourceData, meta interface{}, app cfapi.CCApp, update bool, restart bool, restage bool) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()
	rm := session.RouteManager()

	app.ID = d.Id()

	if update || restart || restage {
		// push any updates to CF, we'll do any restage/restart later
		var err error
		if app, err = am.UpdateApp(app); err != nil {
			return err
		}
		setAppArguments(app, d)
	}

	// update the application's service bindings (the necessary restage is dealt with later)
	if d.HasChange("service_binding") {

		old, new := d.GetChange("service_binding")
		session.Log.DebugMessage("Old service bindings state:: %# v", old)
		session.Log.DebugMessage("New service bindings state:: %# v", new)

		bindingsToDelete, bindingsToAdd := getListChangedSchemaLists(old.([]interface{}), new.([]interface{}))
		session.Log.DebugMessage("Service bindings to be deleted: %# v", bindingsToDelete)
		session.Log.DebugMessage("Service bindings to be added: %# v", bindingsToAdd)

		if err := removeServiceBindings(bindingsToDelete, am, session.Log); err != nil {
			return err
		}

		if added, err := addServiceBindings(app.ID, bindingsToAdd, am, session.Log); err != nil {
			return err
		} else if len(added) > 0 {
			if new != nil {
				for _, b := range new.([]interface{}) {
					bb := b.(map[string]interface{})

					for _, a := range added {
						if bb["service_instance"] == a["service_instance"] {
							bb["binding_id"] = a["binding_id"]
							bb["credentials"] = a["credentials"]
							break
						}
					}
				}
				d.Set("service_binding", new)
			}
		}
		restage = true
	}

	if d.HasChange("route") {
		old, new := d.GetChange("route")

		var (
			oldRouteConfig, newRouteConfig map[string]interface{}
		)

		oldA := old.([]interface{})
		if len(oldA) == 1 {
			oldRouteConfig = oldA[0].(map[string]interface{})
		} else {
			oldRouteConfig = make(map[string]interface{})
		}
		newA := new.([]interface{})
		if len(newA) == 1 {
			newRouteConfig = newA[0].(map[string]interface{})
		} else {
			newRouteConfig = make(map[string]interface{})
		}

		for _, r := range []string{
			"default_route",
			"stage_route",
			"live_route",
		} {
			if _, err := validateRoute(newRouteConfig, r, app.ID, rm); err != nil {
				return err
			}
			if mappingID, err := updateMapping(oldRouteConfig, newRouteConfig, r, app.ID, rm); err != nil {
				return err
			} else if len(mappingID) > 0 {
				newRouteConfig[r+"_mapping_id"] = mappingID
			}
		}
	}

	binaryUpdated := false // check if we need to update the application's binary
	if d.HasChange("url") || d.HasChange("git") || d.HasChange("github_release") || d.HasChange("add_content") {

		var (
			v  interface{}
			ok bool

			appPath string

			addContent []map[string]interface{}
		)

		if appPathCalc, err := prepareApp(app, d, session.Log); err != nil {
			return err
		} else {
			appPath = appPathCalc
		}

		if v, ok = d.GetOk("add_content"); ok {
			addContent = getListOfStructs(v)
		}

		if err := am.UploadApp(app, appPath, addContent); err != nil {
			return err
		}
		binaryUpdated = true
	}

	// now that all of the reconfiguration is done, we can deal doing a restage or restart, as required
	timeout := time.Second * time.Duration(d.Get("timeout").(int))

	// check the package state of the application after binary upload
	var curApp cfapi.CCApp
	var readErr error
	if curApp, readErr = am.ReadApp(app.ID); readErr != nil {
		return readErr
	}
	if binaryUpdated {
		if *curApp.PackageState != "PENDING" {
			// if it's not already pending, we need to restage
			restage = true
		} else {
			// uploading the binary flagged the app for restaging,
			// but we need to restart in order to force that to happen now
			// (this is how the CF CLI does this)
			restage = false
			restart = true
		}
	}

	if restage {
		if err := am.RestageApp(app.ID, timeout); err != nil {
			return err
		}
		if *curApp.State == "STARTED" {
			// if the app was running before the restage when wait for it to start again
			if err := am.WaitForAppToStart(app, timeout); err != nil {
				return err
			}
		}
	} else if restart && !d.Get("stopped").(bool) { // only run restart if the final state is running
		if err := am.StopApp(app.ID, timeout); err != nil {
			return err
		}
		if err := am.StartApp(app.ID, timeout); err != nil {
			return err
		}
	}

	// now set the final started/stopped state, whatever it is
	if d.HasChange("stopped") {
		if d.Get("stopped").(bool) {
			if err := am.StopApp(app.ID, timeout); err != nil {
				return err
			}
		} else {
			if err := am.StartApp(app.ID, timeout); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceAppDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()
	rm := session.RouteManager()

	if v, ok := d.GetOk("service_binding"); ok {
		if err = removeServiceBindings(getListOfStructs(v), am, session.Log); err != nil {
			return
		}
	}
	if v, ok := d.GetOk("route"); ok {

		routeConfig := v.([]interface{})[0].(map[string]interface{})

		for _, r := range []string{
			"default_route_mapping_id",
			"stage_route_mapping_id",
			"live_route_mapping_id",
		} {
			if v, ok := routeConfig[r]; ok {
				mappingID := v.(string)
				if len(mappingID) > 0 {
					if err = rm.DeleteRouteMapping(v.(string)); err != nil {
						if !strings.Contains(err.Error(), "status code: 404") {
							return
						}
						err = nil
					}
				}
			}
		}
	}
	err = am.DeleteApp(d.Id(), false)
	if err = am.DeleteApp(d.Id(), false); err != nil {
		if strings.Contains(err.Error(), "status code: 404") {
			session.Log.DebugMessage(
				"Application with ID '%s' does not exist. App resource will be deleted from state",
				d.Id())
		} else {
			session.Log.DebugMessage(
				"App resource will be deleted from state although deleting app with ID '%s' returned an error: %s",
				d.Id(), err.Error())
		}
	}
	return nil
}

func setAppArguments(app cfapi.CCApp, d *schema.ResourceData) {

	d.Set("name", app.Name)
	d.Set("space", app.SpaceGUID)
	if app.Instances != nil || IsImportState(d) {
		d.Set("instances", app.Instances)
	}
	if app.Memory != nil || IsImportState(d) {
		d.Set("memory", app.Memory)
	}
	if app.DiskQuota != nil || IsImportState(d) {
		d.Set("disk_quota", app.DiskQuota)
	}
	if app.StackGUID != nil || IsImportState(d) {
		d.Set("stack", app.StackGUID)
	}
	if app.Buildpack != nil || IsImportState(d) {
		d.Set("buildpack", app.Buildpack)
	}
	if app.Command != nil || IsImportState(d) {
		d.Set("command", app.Command)
	}
	if app.EnableSSH != nil || IsImportState(d) {
		d.Set("enable_ssh", app.EnableSSH)
	}
	if app.HealthCheckHTTPEndpoint != nil || IsImportState(d) {
		d.Set("health_check_http_endpoint", app.HealthCheckHTTPEndpoint)
	}
	if app.HealthCheckType != nil || IsImportState(d) {
		d.Set("health_check_type", app.HealthCheckType)
	}
	if app.HealthCheckTimeout != nil || IsImportState(d) {
		d.Set("health_check_timeout", app.HealthCheckTimeout)
	}
	if app.Environment != nil || IsImportState(d) {
		d.Set("environment", app.Environment)
	}

	d.SetPartial("timeout")
	d.Set("stopped", *app.State != "STARTED")

	ports := []interface{}{}
	for _, p := range *app.Ports {
		ports = append(ports, p)
	}
	d.Set("ports", schema.NewSet(resourceIntegerSet, ports))
}

func prepareApp(app cfapi.CCApp, d *schema.ResourceData, log *cfapi.Logger) (path string, err error) {

	if v, ok := d.GetOk("url"); ok {
		url := v.(string)

		if strings.HasPrefix(url, "file://") {
			path = url[7:]
		} else {

			var (
				resp *http.Response

				in  io.ReadCloser
				out *os.File
			)

			if out, err = ioutil.TempFile("", "cfapp"); err != nil {
				return "", err
			}

			log.UI.Say("Downloading application %s from url %s.", terminal.EntityNameColor(app.Name), url)

			if resp, err = http.Get(url); err != nil {
				return "", err
			}
			in = resp.Body
			if _, err = io.Copy(out, in); err != nil {
				return "", err
			}
			if err = out.Close(); err != nil {
				return "", err
			}

			path = out.Name()
		}

	} else {
		log.UI.Say("Retrieving application %s source / binary.", terminal.EntityNameColor(app.Name))

		var repository repo.Repository
		if repository, err = getRepositoryFromConfig(d); err != nil {
			return path, err
		}

		if _, ok := d.GetOk("github_release"); ok {
			path = filepath.Dir(repository.GetPath())
		} else {
			path = repository.GetPath()
		}
	}
	if err != nil {
		return "", err
	}

	log.UI.Say("Application downloaded to: %s", path)
	return path, nil
}

func validateRoute(routeConfig map[string]interface{}, route string, appID string, rm *cfapi.RouteManager) (routeID string, err error) {

	if v, ok := routeConfig[route]; ok {

		routeID = v.(string)

		var mappings []map[string]interface{}
		if mappings, err = rm.ReadRouteMappingsByRoute(routeID); err == nil && len(mappings) > 0 {
			if len(mappings) == 1 {
				if app, ok := mappings[0]["app"]; ok && app == appID {
					return routeID, err
				}
			}
			err = fmt.Errorf(
				"route with id %s is already mapped. routes specificed in the 'route' argument can only be mapped to one 'cf_app' resource",
				routeID)
		}
	}
	return routeID, err
}

func updateMapping(
	old map[string]interface{},
	new map[string]interface{},
	route, appID string, rm *cfapi.RouteManager) (mappingID string, err error) {

	var (
		oldRouteID, newRouteID string
	)

	if v, ok := old[route]; ok {
		oldRouteID = v.(string)
	}
	if v, ok := new[route]; ok {
		newRouteID = v.(string)
	}

	if oldRouteID != newRouteID {
		if len(oldRouteID) > 0 {
			if v, ok := old[route+"_mapping_id"]; ok {
				if err = rm.DeleteRouteMapping(v.(string)); err != nil {
					return "", err
				}
			}
		}
		if len(newRouteID) > 0 {
			if mappingID, err = rm.CreateRouteMapping(newRouteID, appID, nil); err != nil {
				return "", err
			}
		}
	}
	return mappingID, err
}

func addServiceBindings(
	id string,
	add []map[string]interface{},
	am *cfapi.AppManager,
	log *cfapi.Logger) (bindings []map[string]interface{}, err error) {

	var (
		serviceInstanceID, bindingID string
		params                       *map[string]interface{}

		credentials        map[string]interface{}
		bindingCredentials map[string]interface{}
	)

	for _, b := range add {
		serviceInstanceID = b["service_instance"].(string)
		params = nil
		if v, ok := b["params"]; ok {
			vv := v.(map[string]interface{})
			params = &vv
		}
		if bindingID, bindingCredentials, err = am.CreateServiceBinding(id, serviceInstanceID, params); err != nil {
			return bindings, err
		}
		b["binding_id"] = bindingID

		credentials = b["credentials"].(map[string]interface{})
		for k, v := range normalizeMap(bindingCredentials, make(map[string]interface{}), "", "_") {
			credentials[k] = v
		}

		bindings = append(bindings, b)
		log.DebugMessage("Created binding with id '%s' for service instance '%s'.", bindingID, serviceInstanceID)
	}
	return bindings, nil
}

func removeServiceBindings(delete []map[string]interface{},
	am *cfapi.AppManager, log *cfapi.Logger) error {

	for _, b := range delete {

		serviceInstanceID := b["service_instance"].(string)
		bindingID := b["binding_id"].(string)

		if len(bindingID) > 0 {
			log.DebugMessage("Deleting binding with id '%s' for service instance '%s'.", bindingID, serviceInstanceID)
			if err := am.DeleteServiceBinding(bindingID); err != nil {
				return err
			}
		} else {
			log.DebugMessage("Ignoring binding for service instance '%s' as no corresponding binding id was found.", serviceInstanceID)
		}
	}
	return nil
}
