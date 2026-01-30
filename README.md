
---

# 🛠️ AI Issue Resolution & Doc-Gap Analyzer

An intelligent system that intercepts incoming customer issues, matches them against a historical knowledge base using vector embeddings, and provides instant solutions. If no match is found, it intelligently routes the task to a human engineer.

---

## 📸 System in Action

### User Interface (Client)

The CLI interface allows users to search for functions or post new issues. The AI immediately checks for existing resolutions.

### Backend Services (Server)

The backend manages the vector database and coordinates between the Go-based API and the Python-based AI logic.

---

## 🛠️ Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/SP1D3R-P/ResolutionSystem.git
cd ResolutionSystem

```

### 2. Build the Project

Ensure you have `make` installed on your system.

```bash
make build
mkdir client\proto
mkdir server\PyRPC\proto
make compile
```

### 3. Environment Configuration

Create a `.env` file in the root directory and configure your ports and database path:

```env
GO_PORT=50000
PY_PORT=50001

VECTOR_DB_PATH='./db/vector_db'

```

---

## 🖥️ Usage

### Running the Server

Start the backend services (API and Vector Database handler):

```bash
make server

```

### Running the Client

Launch the user interface to report or search for issues:

```bash
make user

```


---
