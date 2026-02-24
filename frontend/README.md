# Frontend (SPA)

Quick notes for running and building the frontend, and how to apply backend migrations from the host when needed.

## Run locally (dev)

From the `frontend` folder:

```bash
cd frontend
npm install
npm start
```

The project uses Create React App; in Docker the app is served via the `frontend` service / nginx.

## Build & run with Docker Compose

Build and start services (from repo root):

```bash
docker compose build frontend
docker compose up -d frontend
```

If you changed frontend code and want the new bundle served, rebuild (`docker compose build frontend`) and restart.

## Run backend migrations from host (helpful during frontend testing)

The frontend sometimes needs sample data (channels, demo users). To ensure DB seeders are applied, run the migration commands (adjust `postgres` service name if different):

```bash
docker compose exec -T postgres psql -U sociomile -d sociomile_db < backend/migrations/postgres_init.sql
```

## Useful endpoints for the frontend

- `POST /api/v1/auth/login` — authenticate and get JWT (frontend stores token in `authStore`)
- `GET /api/v1/channels` — used by `ChannelSimulatorPage`
- `POST /api/v1/channel/webhook` — simulator posts here to create conversations/messages
- `GET /api/v1/conversations` — list conversations
- `GET /api/v1/conversations/:id` — conversation + messages (used in `ConversationDetailPage`)
- `POST /api/v1/conversations/:id/messages` — send/append messages

## Troubleshooting

- If the Conversation detail page shows a blank screen, open DevTools Console and Network and check `/api/v1/conversations/:id` response — some DB time fields are returned as objects from the API and the frontend normalizes them in `ConversationDetailPage.tsx` and `ConversationsPage.tsx`.
- When using `curl` from macOS zsh, prefer using files for JSON bodies or use a short Python script to avoid quoting issues.
# Getting Started with Create React App

This project was bootstrapped with [Create React App](https://github.com/facebook/create-react-app).

## Available Scripts

In the project directory, you can run:

### `npm start`

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.\
You will also see any lint errors in the console.

### `npm test`

Launches the test runner in the interactive watch mode.\
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `npm run build`

Builds the app for production to the `build` folder.\
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.\
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.

### `npm run eject`

**Note: this is a one-way operation. Once you `eject`, you can’t go back!**

If you aren’t satisfied with the build tool and configuration choices, you can `eject` at any time. This command will remove the single build dependency from your project.

Instead, it will copy all the configuration files and the transitive dependencies (webpack, Babel, ESLint, etc) right into your project so you have full control over them. All of the commands except `eject` will still work, but they will point to the copied scripts so you can tweak them. At this point you’re on your own.

You don’t have to ever use `eject`. The curated feature set is suitable for small and middle deployments, and you shouldn’t feel obligated to use this feature. However we understand that this tool wouldn’t be useful if you couldn’t customize it when you are ready for it.

## Learn More

You can learn more in the [Create React App documentation](https://facebook.github.io/create-react-app/docs/getting-started).

To learn React, check out the [React documentation](https://reactjs.org/).
