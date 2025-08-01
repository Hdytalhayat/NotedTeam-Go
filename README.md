# NotedTeam - Backend Server (Go)

![Go & Gin](https://raw.githubusercontent.com/gin-gonic/logo/master/color.png)

This is the backend server repository for **NotedTeam**, a high-performance RESTful API built with **Go (Golang)** and the **Gin** framework. It handles all business logic, database management, authentication, and real-time communication for the [NotedTeam mobile app (Flutter)]([Link to Your Frontend Repository Here]).

---

## ✨ Features & Architecture

- **RESTful API**: Provides well-structured endpoints for all CRUD operations (Create, Read, Update, Delete).
- **JWT-Based Authentication**: Secures endpoints using JSON Web Tokens, ensuring only authenticated users can access their data.
- **Real-Time State Management**: Uses **WebSocket** to push instant updates to all connected clients within the same team.
- **Layered Architecture**: Code is organized into separate layers (controllers, models, middlewares, utils) for readability and maintainability.
- **ORM & Migrations**: Utilizes **GORM** for safe database interaction and automatic schema migrations.
- **Input Validation**: Leverages Gin’s built-in validation features to ensure data integrity.
- **Configuration Management**: Uses environment variables (`.env`) to handle sensitive configuration such as database credentials and secret keys.
- **Security Features**:
  - Passwords are hashed using **Bcrypt**.
  - Email verification and password reset workflows using tokens.
  - Authorization middleware to restrict access based on team membership or ownership.

## 🚀 Tech Stack

- **Language**: **Go (Golang)** (version 1.18+)
- **Web Framework**: **Gin Web Framework**
- **Database**: **MySQL**
- **ORM**: **GORM**
- **WebSocket**: **Gorilla WebSocket**
- **Authentication**: **JWT-Go v5**
- **Email**: **Gomail v2**
- **Environment Variables**: **Godotenv**
- **Deployment**: **Railway.app** via **Nixpacks**

## 📂 Project Structure

```

notedteam-backend/
├── config/         # Configuration files (e.g., DB connection)
├── controllers/    # Business logic for handling HTTP & WebSocket requests
├── middlewares/    # Middleware for authentication & authorization
├── models/         # GORM structs representing DB schema
├── utils/          # Helper functions (mailer, token generators)
├── ws/             # Hub logic for WebSocket connection management
├── .env.example    # Example environment variables file
├── go.mod          # Go dependency management
├── main.go         # Application entry point (initializes server & routes)
└── nixpacks.toml   # Build config for deployment on Railway

````

## 📡 API Endpoints

All `/api` routes require the `Authorization: Bearer <token>` header.

### Auth
- `POST /auth/register`: Register a new user & send verification email.
- `POST /auth/login`: Authenticate a user & return JWT token.
- `GET /auth/verify`: Endpoint visited from email to verify account.
- `POST /auth/forgot-password`: Start password reset process.

### Teams
- `GET /api/teams`: Get all teams the user is a member of.
- `POST /api/teams`: Create a new team.
- `GET /api/teams/:teamId`: Get team details, including members (requires membership).
- `PUT /api/teams/:teamId`: Update team name (requires ownership).
- `DELETE /api/teams/:teamId`: Delete team (requires ownership).
- `POST /api/teams/:teamId/invite`: Invite another user to join the team.

### Todos
- `GET /api/teams/:teamId/todos`: Get all to-dos in a team.
- `POST /api/teams/:teamId/todos`: Create a new to-do.
- `PUT /api/teams/:teamId/todos/:todoId`: Update a to-do.
- `DELETE /api/teams/:teamId/todos/:todoId`: Delete a to-do.

### Invitations
- `GET /api/invitations`: Get all pending invitations for the current user.
- `POST /api/invitations/:invitationId/respond`: Accept or decline an invitation.

### WebSocket
- `GET /api/ws/teams/:teamId`: Upgrade to WebSocket connection to receive real-time updates.

## 🏁 Getting Started

### Prerequisites

- **Go**: Version 1.18 or newer.
- **MySQL**: Running MySQL server.
- **SMTP Account**: Required for email verification and password reset features.

### Installation & Running

1. **Clone this repository:**
    ```bash
    git clone [Link to Your Backend Repository Here]
    cd notedteam-backend
    ```

2. **Set Up the Database**:
    - Ensure your MySQL server is running.
    - Create a new database:
      ```sql
      CREATE DATABASE notedteam_db;
      ```

3. **Configure Environment Variables**:
    - Copy `.env.example` to `.env`:
      ```bash
      cp .env.example .env
      ```
    - Fill in the required values:
        - `DB_USER`, `DB_PASSWORD`, `DB_NAME`, etc.
        - `JWT_SECRET` (a strong random string).
        - Your SMTP credentials.

4. **Install Dependencies**:
    ```bash
    go mod tidy
    ```
    or
    ```bash
    go mod download
    ```

5. **Run the Server**:
    ```bash
    go run main.go
    ```
    The server will start, perform DB migrations, and listen on `http://localhost:8080`.
