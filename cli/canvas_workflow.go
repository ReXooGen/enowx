package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/enowx/enowx/proto"
)

func (r runner) canvasWorkflow(command string, args []string) error {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	payload := fs.String("json", "{}", "JSON settings")
	pos, err := parseAll(fs, args)
	if err != nil {
		return err
	}
	if command == "library" {
		var p proto.CanvasLibraryParams
		if err := json.Unmarshal([]byte(*payload), &p); err != nil {
			return err
		}
		p.Session = os.Getenv("ENOWX_SESSION")
		result, err := r.c.CanvasLibrary(p)
		if err != nil {
			return err
		}
		return json.NewEncoder(r.out).Encode(result)
	}
	if len(pos) != 1 {
		return fmt.Errorf("usage: enowx-app canvas workflow CANVAS --json EDIT")
	}
	c, err := r.findCanvas(pos[0])
	if err != nil {
		return err
	}
	var edit proto.CanvasEdit
	if err := json.Unmarshal([]byte(*payload), &edit); err != nil {
		return err
	}
	if edit.Op != proto.CanvasOpBatch && edit.Op != proto.CanvasOpTransition {
		return fmt.Errorf("canvas: workflow edits take elements.batch or frame.transition")
	}
	edit.ID, edit.Session = c.ID, os.Getenv("ENOWX_SESSION")
	result, err := r.c.EditCanvas(edit)
	if err != nil {
		return err
	}
	return json.NewEncoder(r.out).Encode(result)
}

var workflowTools = []mcpTool{
	{name: "canvas_library", description: "List, save, insert or remove reusable HTML components on this host. Save with an existing component ID updates its linked instances. Inline styling or shared destination tokens are recommended.", props: map[string]mcpProp{"settings": {"object", "CanvasLibraryParams: op list/save/insert/remove; component {id,name,html}; canvas,frame,target,into"}}, args: func(a map[string]any) []string {
		data, _ := json.Marshal(a["settings"])
		if a["settings"] == nil {
			data = []byte(`{}`)
		}
		return []string{"library", "--json", string(data)}
	}},
	{name: "canvas_workflow_edit", description: "Arrange, group, lock or move multiple shapes in one undoable edit, or set a prototype transition. Locked shapes must first be unlocked.", props: map[string]mcpProp{"canvas": canvasProp, "edit": {"object", "CanvasEdit: op elements.batch with targets and action, dx/dy; or frame.transition with transition {element,target,trigger,effect,duration,easing}; optional frame"}}, required: []string{"canvas", "edit"}, args: func(a map[string]any) []string {
		data, _ := json.Marshal(a["edit"])
		return []string{"workflow", str(a, "canvas"), "--json", string(data)}
	}},
}
