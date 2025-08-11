# LogLens - The Web-Based Log Inspector

LogLens is a **lightweight, real-time log analysis tool** built in Go.  
It provides a clean web interface to dynamically load, view, and filter structured (JSON) log files — turning your browser into a powerful log inspector.  

It’s designed as a **fast, local alternative** to heavy cloud-based log analysis platforms for **quick debugging and investigation**.

---


---

## 🚀 Features

- **Dynamic Web UI** – Modern, clean interface built with vanilla HTML, CSS, and JavaScript.
- **Drag & Drop Upload** – Drop your `.jsonl`, `.log`, or `.txt` file directly into the browser.
- **Real-Time Processing** – Backend in Go streams parsed logs via **WebSockets**.
- **Interactive Panes** – Two-pane layout: summary list + detailed JSON view.
- **Live Filtering** – Instant filtering of thousands of log lines.
- **High Performance** – Go concurrency ensures smooth, non-blocking parsing & streaming.
- **Zero Dependencies for Users** – Runs as a single Go executable.

---

## 💡 Why LogLens?

In a world of complex log management systems like **ELK Stack** or **Datadog**, sometimes you just think:

> _“I have a log file. I want to quickly search through it — locally.”_

LogLens gives you:
- **Simplicity** – No setup, no DB, no configs. Just run & open in browser.
- **Speed** – Go’s performance and concurrency make it ideal for text parsing.
- **Privacy** – All logs are processed locally, never sent to the internet.

---

## 🛠️ Tech Stack

- **Go** (`net/http`) – Serving web content & APIs.
- **Gorilla/WebSocket** – Real-time, bidirectional communication.
- **Goroutines & Channels** – Concurrent file processing & hub-based WS management.
- **HTML/CSS/JS** – Clean, dependency-free frontend.

---

## 📦 Getting Started

### Prerequisites
- **Go** v1.18+ installed.
- A web browser (Chrome, Firefox, Safari).

### Installation & Running

```bash
# Clone repository
git clone https://github.com/your-username/loglens.git
cd loglens

# Download Go dependencies
go mod tidy

# Run the server
go run .

# Open in your browser
http://localhost:8080
