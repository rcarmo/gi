package tui

import (
	"context"
	"fmt"

	"github.com/rcarmo/gi/internal/inference"
)

func (c *chatTUI) modelDefaults() inference.SessionModelChoice {
	if c.sessionModelDefaults == nil {
		c.sessionModelDefaults = &[3]string{c.cfg.DefaultModel, c.cfg.DefaultProvider, c.cfg.DefaultThinkingLevel}
	}
	d := c.sessionModelDefaults
	return inference.SessionModelChoice{Model: d[0], Provider: d[1], Thinking: d[2]}
}

func (c *chatTUI) sessionModelCatalogue() []inference.ModelOption {
	defaults := c.modelDefaults()
	_, options := inference.ListRuntimeOptions(defaults.Provider, defaults.Model, c.cfg.EnabledModels)
	return options
}

func (c *chatTUI) restoreSessionModel(state map[string]any) {
	choice := inference.SessionModel(state, c.modelDefaults())
	c.cfg.DefaultModel, c.cfg.DefaultProvider, c.cfg.DefaultThinkingLevel = choice.Model, choice.Provider, choice.Thinking
}

func (c *chatTUI) chooseSessionModel(requested string) error {
	if c.store == nil || c.sessionID == "" {
		return fmt.Errorf("no active session")
	}
	session, err := inference.SelectSessionModel(context.Background(), c.store, c.sessionID, c.sessionModelCatalogue(), requested)
	if err != nil {
		return err
	}
	idleStatus := fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.restoreSessionModel(session.State)
	if c.status == idleStatus {
		c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	}
	return nil
}

func (c *chatTUI) sessionModelLabel() string {
	return (inference.SessionModelChoice{Model: c.cfg.DefaultModel, Provider: c.cfg.DefaultProvider}).Label()
}
