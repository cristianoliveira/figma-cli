---
title: "Rate Limiting"
tags: ["figma-api", "resources", "api-docs", "rate-limiting"]
---
# Rate Limiting

Rate limits are multi-dimensional based on:
1. User seat type (View/Collab vs Dev/Full)
2. API endpoint tier (Tier 1, 2, or 3)
3. Resource plan (Starter, Professional, Organization, Enterprise)

| API Tier | Seat Type | Starter | Professional | Organization | Enterprise |
|----------|-----------|---------|--------------|--------------|------------|
| Tier 1   | View/Collab | Up to 6/month | Up to 6/month | Up to 6/month | Up to 6/month |
|          | Dev/Full  | 10/min  | 15/min       | 20/min       | 20/min     |
| Tier 2   | View/Collab | Up to 5/min | Up to 5/min | Up to 5/min | Up to 5/min |
|          | Dev/Full  | 25/min  | 50/min       | 100/min      | 100/min    |
| Tier 3   | View/Collab | Up to 10/min | Up to 10/min | Up to 10/min | Up to 10/min |
|          | Dev/Full  | 50/min  | 100/min      | 150/min      | 150/min    |

**Key Points**:
- Uses leaky bucket algorithm
- 429 errors include `Retry-After` header
- View/Collab seats have significantly lower limits
- Dev/Full seats have higher limits (up to 150/min for Tier 3 Enterprise)

