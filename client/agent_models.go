package client

import "github.com/enowx/enowx/proto"

// AgentCommands lists commands on the connected host before starting a session.
func (c *Client) AgentCommands(p proto.AgentModelsParams) (proto.AgentCommands, error) {
	var result proto.AgentCommands
	err := c.Call(proto.MethodAgentCommands, p, &result)
	return result, err
}

// AgentModels lists models on the connected host, using the CLI's registry.
func (c *Client) AgentModels(p proto.AgentModelsParams) (proto.AgentModels, error) {
	var result proto.AgentModels
	err := c.Call(proto.MethodAgentModels, p, &result)
	return result, err
}
