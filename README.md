# Dead by Daylight Character Tracker

This is a small Go web application designed to help you keep track of the various killer characters you've played against in the game Dead by Daylight.

---

## Features

* **Character Tracking**: Log and view details about the Dead by Daylight characters you encounter.
* **Web Interface**: Access and manage your character data through a simple web browser.
* **Lightweight**: Built with Go, making it efficient and easy to deploy.

---

## Getting Started

### Prerequisites

You'll need the following installed on your machine:

* **Go (1.23.5 or newer recommended)**: You can download it from [golang.org/doc/install](https://golang.org/doc/install).

### Installation

1.  **Clone the repository:**

    ```bash
    git clone [https://github.com/Ryan-Har/dbd_history.git](https://github.com/Ryan-Har/dbd_history.git)
    cd dbd_history
    ```

2.  **Run the application:**

    ```bash
    go run .
    ```

    The application will typically start on `http://localhost:8080`.
    The application uses SQLite for persistence, typically this is in the location `./db/characters.db`

---

## Docker Support

This application also comes with a `Dockerfile` for easy containerization.

1.  **Build the Docker image:**

    ```bash
    docker build -t dbd-tracker .
    ```

2.  **Run the Docker container:**

    ```bash
    docker run -p 8080:8080 dbd-tracker
    ```

    Your application will be accessible via `http://localhost:8080` in your browser.

---

## Contributing

Feel free to fork the repository, make improvements, and submit pull requests.

---
