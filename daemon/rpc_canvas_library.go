package daemon

import "github.com/enowx/enowx/proto"

func canvasLibraryHandler(c *conn, p proto.CanvasLibraryParams) (any, error) {
	if p.Session != "" && p.Op != "list" && p.Op != "" {
		list, err := c.srv.canvases.List()
		if err != nil {
			return nil, err
		}
		for _, cv := range list {
			if c.srv.canvasState.isPaused(cv.ID) {
				return nil, errCanvasPaused
			}
		}
	}
	result, err := c.srv.canvases.Library(p)
	for _, cv := range result.Canvases {
		c.srv.broadcast(proto.EventCanvasChanged, proto.CanvasChange{Canvas: c.srv.withState(cv), Op: "component.update"})
	}
	// Broadcast each linked update separately; the response only needs this canvas.
	canvases := result.Canvases
	result.Canvases = nil
	for _, cv := range canvases {
		if cv.ID == p.Canvas {
			result.Canvases = append(result.Canvases, cv)
		}
	}
	return result, err
}
