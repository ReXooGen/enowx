# Canvas workflow

Open a canvas and choose **Canvas tools** in the toolbar. The workflow panel has six sections. These controls use the host's daemon for both This PC and VPS canvases; an older daemon needs updating before it can store the new operations.

The Canvas agent panel offers **Chat / Terminal** for agents with transcript support, including OMP. Terminal shows the same running session's live CLI, including progress, errors and prompts that may not appear in its transcript. Switching views does not restart the agent.

OMP chat separates the thoughts it records from tool calls and their results. Tool results attach to their calls, failures remain marked, and embedded images show small previews. Images are decoded on the host into bounded thumbnails; unsupported or oversized images show a placeholder. This does not enable vision for a model that does not support it. Remote sessions need the updated daemon to send these details.

## Agent controls

The **Provider** and **Model** dropdowns appear directly in OMP's Canvas composer. The connected host runs `omp models --json` in the session's folder (or the canvas agent's folder before starting). The dropdowns offer only that registry's chat models; the refresh icon reads configuration changes. Selecting a model configures the next conversation or switches an idle session. Finish the current task before changing it. Only model identifiers and display names cross the protocol; provider credentials stay on the host.

Recorded **Thinking / Thoughts** has its own visible preview and expandable details, independent of the folded tool history.

While working, recorded thoughts stay open and newest first; **Running tools** distinguishes pending calls from the thinking phase. Only thoughts recorded by OMP appear. Slash commands show common OMP commands immediately, then add the CLI's discovered commands, skills and extensions in the background, even before a conversation starts.

The activity pill inside the Canvas follows the session's Working and Needs input states, including time spent thinking or running tools before the first canvas edit.

## Library

Select an HTML mockup element, enter a name and choose **Save selected component**. Insert a saved component into another design. **Update** replaces the saved definition from the selection and updates its linked instances across canvases on this host. Each insertion gets unique IDs, including label and accessibility references. Removing a definition keeps existing instances.

Components contain one HTML root. Use inline styles or design tokens available in the destination; page-wide styles are not copied. Imported image assets are embedded when saving. The library allows up to 2 MB per component and 8 MB in total. Updates are versioned so each affected canvas can be undone separately.

## Review

Choose **Compare** on an older version, then open **Review**. Keep or reject a listed element without restoring the entire canvas. Decisions are recorded. A rejection restores only that element; edits to other elements stay. If the element changed after the compared version, compare again. Agents cannot submit review decisions. Frame-wide changes without an element ID still use the existing version controls.

## Checks

Choose **Check this design** to inspect shape geometry and mockup layouts at 1440, 768 and 390 px. Findings cover overflow, clipped text, text contrast, small controls and overlapping interactive controls. Select a finding to locate its element.

These are heuristics, not an accessibility certification. Decorative overlaps, image backgrounds and dynamic states need manual review. Preview copies do not execute scripts; generated markup and current styles are inspected. Checks examine at most 2,000 nodes per preview and report up to 150 findings.

## Arrange

Select a connector to change **Routing**: Straight, Elbow 90° (automatic), Elbow horizontal or Elbow vertical. New connectors made with Connect use 90° turns. Endpoints follow their shapes, and the same route is used in SVG, HTML, PNG and PDF exports. Routing does not avoid unrelated shapes automatically.

Shift-click shapes to select several, or use **Select all shapes** / Ctrl/⌘+A. Align edges or centers, distribute three or more shapes with equal gaps, group, ungroup, lock or unlock. Dragging a selection moves it together in one undoable edit. Grouped shapes select together. Ctrl/⌘+G groups; Ctrl/⌘+Shift+G ungroups. Locked shapes must be unlocked before changing or deleting them. These operations apply to native shapes; HTML elements retain their existing single-element editor.

## Motion

Select a shape or HTML element, choose a target design, click or hover, none/fade/slide, easing and a duration from 0–3000 ms. Choose **Apply transition**, then switch to **Interact** to try it. Export all designs to preserve navigation between them. HTML previews respect the device's reduced-motion preference. An individual design export has no other designs to navigate to.

## Offline

Export this design or all designs as a single HTML file. Imported assets, public images, fonts, CSS and classic scripts are embedded. The host downloads public HTTP(S) resources without credentials; private resources must first be imported as canvas assets. The exported page blocks runtime network requests.

JavaScript modules, responsive `srcset`, layered CSS imports and runtime APIs need a project build; unsupported resources produce an error instead of a partial export. Limits are 96 downloads, 4 MB per resource, 6 MB total downloaded assets and 8 MB final HTML. Larger designs should be exported from their project.

Agents can use `canvas_library` and `canvas_workflow_edit` through the existing Canvas feature switch. The CLI equivalents are `enowx-app canvas library --json SETTINGS`, `canvas workflow CANVAS --json EDIT` and `canvas export CANVAS --format offline-html --out preview.html`.
