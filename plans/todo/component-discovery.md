# Frontend component discovery

## Problem

`figma components` sounds like design-system inventory but currently lists generic descendants. Frontend developer expects components, instances, variants, component properties, and source references useful for mapping designs to implementation.

## Solution

Make `components` domain-specific. Return only component-related nodes by default with useful variant and instance metadata. Keep generic tree discovery in `find`, `layout`, and `inspect`.

## How

- Define output for COMPONENT, COMPONENT_SET, and INSTANCE nodes.
- Include component IDs, component-set IDs, variant properties, exposed properties, names, node path, and instance/source relationship.
- Add flags for `--kind component|set|instance`, `--name`, and optional raw output.
- Decide whether current generic behavior is removed or moved to explicit `--all-nodes` during active refactor.
- Add tests using component set and detached-instance fixtures.

## Success criteria

- Default output answers “which reusable components does this frame use?”
- Variant mapping is available without raw JSON.
- Command name matches behavior.
