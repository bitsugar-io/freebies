# API Reference

Base URLs:

- Local: `http://localhost:8080/api/v1`
- Production: `https://freebie-api.fly.dev/api/v1`

## Leagues

- `GET /leagues` - List all leagues

## Events

- `GET /events` - List all events
- `GET /events/:id` - Get event by ID

## Users

- `POST /users` - Create user
- `GET /users/:id` - Get user
- `PUT /users/:id/push-token` - Update push token

## Subscriptions

- `GET /users/:id/subscriptions` - List subscriptions
- `POST /users/:id/subscriptions` - Subscribe to event
- `DELETE /users/:id/subscriptions/:eventId` - Unsubscribe

## Active Deals

- `GET /users/:id/active-deals` - List active deals

## Dismissals

- `POST /users/:id/dismissals` - Dismiss a deal
- `DELETE /users/:id/dismissals/:triggeredEventId` - Undo dismissal

## User Stats

- `GET /users/:id/stats` - Get user statistics (deals claimed, subscriptions count)
