package proto

import "time"

const (
	MethodCanvasLibrary = "canvas.library"
	CanvasOpBatch       = "elements.batch"
	CanvasOpReview      = "element.review"
	CanvasOpTransition  = "frame.transition"
)

// CanvasElementReview records a decision about one element between two versions.
type CanvasElementReview struct {
	Frame    string `json:"frame"`
	Element  string `json:"element"`
	Base     int    `json:"base"`
	Head     int    `json:"head"`
	Decision string `json:"decision"`
}

// CanvasTransition describes navigation from an element to another design.
type CanvasTransition struct {
	Element  string `json:"element"`
	Target   string `json:"target"`
	Trigger  string `json:"trigger"`
	Effect   string `json:"effect"`
	Duration int    `json:"duration"`
	Easing   string `json:"easing"`
}

// CanvasLibraryComponent is a host's reusable HTML component.
type CanvasLibraryComponent struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	HTML    string    `json:"html"`
	Updated time.Time `json:"updated"`
}

// CanvasLibraryParams lists, saves, removes or inserts a reusable component.
// Save with Component.ID updates that component and its linked instances.
type CanvasLibraryParams struct {
	Session   string                 `json:"session,omitempty"`
	Op        string                 `json:"op"`
	Component CanvasLibraryComponent `json:"component"`
	Canvas    string                 `json:"canvas,omitempty"`
	Frame     string                 `json:"frame,omitempty"`
	Target    string                 `json:"target,omitempty"`
	Into      string                 `json:"into,omitempty"`
}

// CanvasLibraryResult includes changed canvases so connected apps can refresh.
type CanvasLibraryResult struct {
	Components []CanvasLibraryComponent `json:"components"`
	Canvases   []Canvas                 `json:"canvases,omitempty"`
}
