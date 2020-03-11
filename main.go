package main

import (
	"net/http"
	"os"

	"github.com/Nacdlow/plugin-sdk"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"gopkg.in/macaron.v1"
)

// MCPlugin is an implementation of IgluPlugin
type MCPlugin struct {
	logger hclog.Logger
}

func (g *MCPlugin) OnLoad() error {
	go startWebServer(g)
	return nil
}

type LightGroup struct {
	Id     string
	Status bool
}

var (
	registeredLightGroups []LightGroup
	serverStarted         = false
)

func startWebServer(g *MCPlugin) {
	g.logger.Debug("Starting Minecraft Simulation Bridge server on port :25564")
	m := macaron.Classic()
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:  "DO_NOT_USE",
		IndentJSON: true,
	}))
	m.Get("/", func() string {
		return "Hello! This is the Minecraft Simulation plugin's bridge server"
	})
	m.Get("/register_light_group/:id", func(ctx *macaron.Context) {
		for _, v := range registeredLightGroups {
			if v.Id == ctx.Params("id") {
				return // Already registered
			}
		}
		registeredLightGroups = append(registeredLightGroups, LightGroup{
			Id:     ctx.Params("id"),
			Status: false,
		})
	})
	m.Get("/get_device_states", func(ctx *macaron.Context) {
		ctx.JSON(http.StatusOK, registeredLightGroups)
	})
	m.Get("/toggle_group_status/:id", func(ctx *macaron.Context) {
		for i, v := range registeredLightGroups {
			if v.Id == ctx.Params("id") {
				registeredLightGroups[i].Status = !registeredLightGroups[i].Status
				return
			}
		}
	})
	g.logger.Error(http.ListenAndServe("0.0.0.0:25564", m).Error())
	panic("server cannot be started")
}

func (g *MCPlugin) GetManifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Id:      "minecraft-simulation-iglu-plugin",
		Name:    "Minecraft Simulation",
		Author:  "Nacdlow",
		Version: "v0.1",
	}
}

func (g *MCPlugin) OnDeviceToggle(id string, status bool) error {
	for i, v := range registeredLightGroups {
		if "mc_"+v.Id == id {
			registeredLightGroups[i].Status = status
			return nil
		}
	}
	return nil
}

func (g *MCPlugin) GetDeviceStatus(id string) bool {
	for _, v := range registeredLightGroups {
		if "mc_"+v.Id == id {
			return v.Status
		}
	}
	return false
}

func (g *MCPlugin) GetPluginConfiguration() []sdk.PluginConfig {
	return []sdk.PluginConfig{
		sdk.PluginConfig{
			Title:          "Bridge Server Port",
			Description:    "The port the bridge web server will be hosted on.",
			Key:            "bridge-port",
			Default:        "25564",
			Type:           sdk.NumberValue,
			IsUserSpecific: false,
		},
	}
}

func (g *MCPlugin) OnConfigurationUpdate(config []sdk.ConfigKV) {
}

func (g *MCPlugin) GetAvailableDevices() []sdk.AvailableDevice {
	var devs []sdk.AvailableDevice
	for _, v := range registeredLightGroups {
		devs = append(devs, sdk.AvailableDevice{
			UniqueID:         "mc_" + v.Id,
			ManufacturerName: "Ferret's Hue",
			ModelName:        "Redstone++",
			Type:             0,
		})
	}
	return devs
}

func (g *MCPlugin) GetWebExtensions() []sdk.WebExtension {
	return []sdk.WebExtension{}
}

// DO NOT CHANGE below this line

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "IGLU_PLUGIN",
	MagicCookieValue: "MzlK0OGpIRs",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	test := &MCPlugin{
		logger: logger,
	}

	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"iglu_plugin": &sdk.IgluPlugin{Impl: test},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
