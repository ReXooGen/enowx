package proto

// MethodAgentModels reads an agent's model catalog on its host.
const MethodAgentModels = "agent.models"

// AgentModelsParams selects the same agent and folder as a session or canvas.
type AgentModelsParams struct {
	Agent   string `json:"agent"`
	Session string `json:"session,omitempty"`
	Canvas  string `json:"canvas,omitempty"`
}

// AgentModel contains display identifiers only, never provider credentials.
type AgentModel struct {
	Provider string `json:"provider,omitempty"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Selector string `json:"selector"`
}

// AgentModels distinguishes a live CLI registry from built-in agent choices.
type AgentModels struct {
	Models     []AgentModel `json:"models"`
	Registered bool         `json:"registered"`
}
