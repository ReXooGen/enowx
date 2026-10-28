package agent

// CanvasVariationsRequest is the short request shown by the Variations preset.
const CanvasVariationsRequest = "Create 5 distinct design directions for this design. Preserve its content and meaning; give each direction a clear rationale and review the rendered results."

// CanvasVariationsGuide is the reusable brief for deliberate design alternatives.
const CanvasVariationsGuide = `## Deliberate design variations
For the five-variation preset, create exactly five NEW frames and preserve the source. For another requested count, adapt this process to that count. Follow the user's explicit visual direction first.

1. Read the selected frame, comments and attached project's design tokens, components and references. State the actual purpose, audience, primary task and constraints briefly. Treat missing facts as unknown. Keep the source's language, content, actions, nodes and relationships consistent across alternatives; do not invent features, metrics, testimonials or technical behavior to fill space.

2. Define five visual hypotheses before drawing. Each needs a descriptive title, a sentence explaining the user benefit, and its trade-off. Derive them from this design's task. Explore different information priorities, composition, grouping, density, reading order or interaction emphasis. Each pair must differ on at least two meaningful axes. Keep the product's identity recognizable. Inspect silhouettes at thumbnail scale; replace alternatives that read as duplicates. Avoid making five palette swaps or applying the same five style labels to every product.

3. Build with restraint. Reuse existing fonts, components and tokens. Without an established system, choose a coherent type hierarchy, spacing rhythm, surface and purposeful accent that fit this content. Use real source copy. Add imagery, decoration or motion only when it explains or supports the task. Avoid automatic hero slogans, repeated icon-card grids, glass panels, purple gradients, oversized rounded containers, fake badges and decorative section numbering. These are choices only when the brief or established project actually calls for them.

4. Match the medium. For diagrams and wireframes, preserve the graph and use native shapes: vary composition, hierarchy, spacing and grouping while keeping labels and connected arrows readable. Do not add gateways, retries or workflow steps unless present in the source. For UI mockups, make the primary action, navigation and relevant states usable. Do not turn a technical diagram into a marketing page.

5. Review the rendered alternatives at full size and as thumbnails. Check text fit, contrast, spacing, alignment, content preservation and connector visibility. Compare the five side by side; fix clipping, hidden arrows and generic repeated compositions before finishing. If rendering is unavailable, say what remains unverified. End with one short line per frame describing its hypothesis and trade-off, then recommend the best fit for the stated task with a concrete reason. Do not claim one direction is objectively best without evidence.`
