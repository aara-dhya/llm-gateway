# Self-Hosted LLM API Gateway & Observability Dashboard

A lightweight, high-performance reverse proxy and observability dashboard for Large Language Models (LLMs). This tool intercepts requests to providers like OpenAI, tracks token usage, calculates costs, and provides a UI to monitor API activity.

## 🛠 Tech Stack
* **Backend:** Go (net/http)
* **Database:** SQLite (Embedded)
* **Frontend:** SvelteKit & Tailwind CSS
* **Infrastructure:** Docker & Kubernetes (Minikube)

## 🚀 Getting Started

### Prerequisites
* Go 1.21+

### Running the Gateway (Development)
1. Clone the repository.
2. Run the Go backend:
   ```bash
   go run main.go