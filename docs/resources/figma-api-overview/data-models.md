---
title: "Data Models"
tags: ["figma-api", "resources", "api-docs", "data-models"]
---
# Data Models

**Node Types:**
- `DOCUMENT` - Root node containing pages
- `CANVAS` (Page) - Top-level page in a file
- `FRAME` - Container with layout
- `GROUP` - Grouped elements
- `SECTION` - Organizational grouping
- `VECTOR` - Vector paths
- `BOOLEAN_OPERATION` - Union, subtract, intersect, exclude
- `STAR`, `LINE`, `ELLIPSE`, `REGULAR_POLYGON`, `RECTANGLE` - Primitive shapes
- `TEXT` - Text elements
- `COMPONENT` - Reusable design element
- `COMPONENT_SET` - Collection of variants
- `INSTANCE` - Instance of a component with overrides

**Design Properties:**
- **Paint**: Fill definitions (solid, gradient, image, video)
- **Layout Constraints**: Positioning behavior within frames
- **Transform**: 2D transformation matrices

**Component System:**
- **Component**: Reusable design element published to team libraries
- **Component Set**: Collection of component variants
- **Instance**: Instance of a component with overrides

**Variable System (Enterprise):**
- **Variable**: Design variable with values per mode
- **Variable Collection**: Group of variables with shared modes
- **Variable Alias**: Reference to a variable value

