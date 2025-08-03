# Target Engine

This project implements a target engine API for managing and serving campaigns based on specific targeting rules. The application is built with performance in mind, prioritizing fast read operations for high-throughput use cases.

## Overview

The assignment has been completed and tested locally, achieving the following response time:

**GET /v1/delivery 200 258.667µs 1754251702335460000**

The application is deployed on Render (Free Tier), which may introduce slight delays due to cold starts. The GitHub repository for this project is available here:

🔗 [https://github.com/Harshi-itaSinha/target-engine](https://github.com/Harshi-itaSinha/target-engine)

## Testing the Endpoint

To test the `/v1/delivery` endpoint, use the following curl command:

```bash
curl -i -X GET "https://target-engine.onrender.com/v1/delivery?app=com.example.finance&country=IN&os=android"
```

## Design and Implementation

The current implementation uses **MongoDB** as the database due to budget constraints, although **DynamoDB** was considered for its high read performance. The design prioritizes fast read operations by storing precomputed and duplicated data, making writes and campaign setup slower to optimize for read-heavy workloads.

### Database Schema

The application uses three MongoDB collections:

1. **Campaign**: Stores details of all campaigns.
2. **Target Rule**: Contains targeting rules associated with campaigns.
3. **Active Campaign Target**: A precomputed collection that stores only active campaigns to enable faster read operations. When a campaign is activated, its data is pushed to this collection. Updates to the `Campaign` or `Target Rule` collections are reflected here.

### Approach

- **Write Optimization**: Campaign creation and rule setup are intentionally slower to ensure data consistency across collections.
- **Read Optimization**: The `Active Campaign Target` collection contains precomputed data for active campaigns, reducing query complexity and improving response times for read operations.
- **Endpoints**: The API exposes endpoints for creating campaigns and rules, in addition to the `/v1/delivery` endpoint for retrieving active campaigns based on parameters like `app`, `country`, and `os`.

## Future Improvements

- Continue updating documentation to improve clarity and completeness.
- Add more features as time permits, such as enhanced filtering or analytics.

## Feedback

For any questions or feedback, please reach out to the project maintainer.