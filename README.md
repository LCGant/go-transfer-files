# Uploader-Files

Uploader-Files is a simple and secure file uploading and sharing tool built with Golang and Gin framework. It allows users to upload files and generate temporary download links for secure sharing.

## Features

- Upload files up to **100MB**
- Generate **temporary** download links with customizable expiration times
- Secure file storage and retrieval
- Simple and intuitive UI for easy file management
- Built using **Golang**, **Gin**, and **GORM**

## Installation

To run the project locally, follow these steps:

### Prerequisites

- **Golang** installed (version 1.18 or higher recommended)
- **MySQL** database setup
- Git (optional for cloning)

### Clone the Repository

```sh
git clone https://github.com/LCGant/go-transfer-files.git
cd go-transfer-files
```

### Configure Database

Before running the application, configure your **MySQL** database connection in `main.go`:

```go
dsn := "root:@tcp(127.0.0.1:3306)/files?charset=utf8mb4&parseTime=True&loc=Local"
```

Modify this according to your MySQL credentials.

### Run the Application

```sh
go run main.go
```

This will start the application on **localhost:8080**.

## Usage

1. Open your browser and go to `http://localhost:8080/`
2. Select **Upload File** and choose a file.
3. Choose the **expiration time** for the file.
4. Click **Upload** to generate a download link.
5. Share the link with others to allow file downloads.

## API Endpoints

### **Upload a File**

```http
POST /Files/upload
```

**Request Parameters:**

- `file`: The file to upload (multipart/form-data)
- `availabilityDuration`: Duration in minutes (integer)

**Response:**

```json
{
  "message": "File uploaded successfully",
  "download_link": "/lokiFiles/download?token=your-token"
}
```

### **Download a File**

```http
GET /Files/download?token=your-token
```

Returns the requested file for download if it hasn't expired.

## Security Considerations

> **Note:** If you plan to deploy this project in a production environment, ensure the following security measures are implemented:

- **XSRF/CSRF Protection:** Prevent cross-site request forgery attacks by implementing robust token-based protection mechanisms.
- **Secure Cookie Handling:** Use attributes like `HttpOnly`, `Secure`, and `SameSite` to protect cookies from being accessed or transmitted in an insecure manner.
- **TLS/HTTPS:** Enable HTTPS to ensure encrypted communication between the server and clients.
- **Environment Variables:** Store sensitive information, such as database credentials, in environment variables instead of hardcoding them in the source code.
- **WebSocket Security:** Review and test WebSocket communications to identify and address potential vulnerabilities.

## Contributing

Pull requests are welcome! If you have suggestions for improving this project, feel free to fork the repository and submit a PR.

## License

This project is open-source and licensed under the **MIT License**.

