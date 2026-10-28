package client

import "github.com/enowx/enowx/proto"

// AgentModels lists models on the connected host, using the CLI's registry.
func (c *Client) AgentModels(p proto.AgentModelsParams) (proto.AgentModels, error) {
	var result proto.AgentModels
	err := c.Call(proto.MethodAgentModels, p, &result)
	return result, err
}
