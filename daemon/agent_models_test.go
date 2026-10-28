package daemon

import (
	"github.com/enowx/enowx/proto"
	"testing"
)

func TestAgentModelCatalogUsesTheConnectedHost(t *testing.T) {
	srv := startServer(t)
	a, b := connect(t, srv, nil), connect(t, srv, nil)
	cv, err := a.CreateCanvas(proto.CanvasCreateParams{Title: "Model catalog"})
	if err != nil {
		t.Fatal(err)
	}
	for _, cl := range []interface {
		AgentModels(proto.AgentModelsParams) (proto.AgentModels, error)
	}{a, b} {
		result, err := cl.AgentModels(proto.AgentModelsParams{Agent: "claude", Canvas: cv.ID})
		if err != nil || len(result.Models) == 0 || result.Registered {
			t.Fatalf("host catalog: %+v %v", result, err)
		}
		if _, err = cl.AgentModels(proto.AgentModelsParams{Agent: "not-a-preset", Canvas: cv.ID}); err == nil {
			t.Fatal("unknown agent accepted")
		}
		if _, err = cl.AgentModels(proto.AgentModelsParams{Agent: "omp", Canvas: "missing"}); err == nil {
			t.Fatal("unknown canvas accepted")
		}
	}
}
