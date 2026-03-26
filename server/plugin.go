package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// client is the Mattermost server API client.
	client *pluginapi.Client

	// router is the HTTP router for handling API requests.
	router *mux.Router

	backgroundJob *cluster.Job
	botUserID     string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	p.router = p.initRouter()

	if err := p.ensureBot(); err != nil {
		return errors.Wrap(err, "failed to ensure bot")
	}

	if err := p.registerCommand(); err != nil {
		return errors.Wrap(err, "failed to register slash command")
	}

	job, err := cluster.Schedule(
		p.API,
		backgroundJobKey,
		cluster.MakeWaitForRoundedInterval(time.Minute),
		p.runJob,
	)
	if err != nil {
		return errors.Wrap(err, "failed to schedule background job")
	}

	p.backgroundJob = job

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	if p.backgroundJob != nil {
		if err := p.backgroundJob.Close(); err != nil {
			p.API.LogError("Failed to close background job", "err", err)
		}
	}
	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	response, err := p.handleCommand(args)
	if err != nil {
		p.API.LogError("Echo Summary command failed", "user_id", args.UserId, "command", args.Command, "err", err)
		return p.ephemeralCommandResponse(fmt.Sprintf("요청을 처리하지 못했습니다: %s", err.Error())), nil
	}
	return response, nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
