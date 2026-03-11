# TCP-Chat (net-cat)

**Description**
This project is a simple TCP chat server written in Go. It lets multiple people connect with a TCP client (for example `nc`, `telnet`, or the included TUI client) and chat in real time. Each message is timestamped, and new users receive the recent chat history when they join.

The main program (`main.go`) starts the server and waits for incoming connections. Each connected user chooses a name, then can send messages to everyone else. The server keeps all state in memory.

**Features**
- Accepts up to 10 clients at the same time.
- Default port is `8989`, but you can choose a different port.
- ASCII art welcome banner with a name prompt.
- Join and leave system messages.
- Timestamped chat messages (`YYYY-MM-DD HH:MM:SS`).
- In‑memory chat history sent to new users when they connect.
- Ignores empty messages (no blank lines in chat).
- Includes an optional terminal UI (TUI) client.
- Tests cover the main behaviors.
- Clear join, message, and leave flow (explained below).

**Requirements**
- Go `1.24.3` or newer (as listed in `go.mod`).
- For the optional TUI client: a terminal that supports cursor control (the TUI uses `gocui`).

**Installation**
1. Install Go.
2. Clone or download this repository.
3. Open a terminal in the project folder.

**Usage**

Start the server (default port `8989`):
```bash
go run .
```

Start the server on a custom port (example `9000`):
```bash
go run . 9000
```

Connect using `nc` (netcat):
```bash
nc localhost 8989
```

Connect using `telnet` (if you prefer):
```bash
telnet localhost 8989
```

Use the built‑in TUI client (connect to a running server):
```bash
go run ./cmd/tui 8989
```

Use the TUI client and start a local server automatically:
```bash
go run ./cmd/tui --with-server
```

Quit the TUI with `Ctrl+C` or `Ctrl+Q`.

**Example**

Example using `nc` (timestamps will match the current time):

Server output:
```
Listening on the port :8989
```

Client 1 (Alice) sees:
```
Welcome to TCP-Chat!
...
[ENTER YOUR NAME]:
```
Alice types:
```
Alice
```

Client 2 (Bob) connects and sees the same banner, then types:
```
Bob
```

Now Alice receives a system message:
```
[2026-03-11 12:34:56][System]: Bob has joined our chat.
```

Bob sends a message:
```
Hello everyone!
```

Alice receives:
```
[2026-03-11 12:34:58][Bob]: Hello everyone!
```

Note: the server does **not** echo your own message back to you. If you use `nc` or `telnet`, you will not see your own line after pressing Enter. The TUI client prints your own messages locally so you can see them.

**How the Program Works (Step by Step)**

This section walks through the full flow, from starting the program to clients joining, sending messages, and leaving. It is written in the exact order the code runs.

**1. Program start**
- `main.go` is the entry point.
- It reads the optional port argument.
- If you give more than one argument, it prints a usage message and exits.
- It calls `server.Start(port)` to start listening.

**2. Server startup**
- `server.Start()` calls `server.StartWithOptions()`.
- If no port was provided, it uses `service.DefaultPort` (`8989`).
- The server creates a TCP listener with `net.Listen("tcp", ":"+port)`.
- It prints `Listening on the port :<port>`.
- It creates the in‑memory chat server with `service.NewServer(maxClients)`.
- Then it waits forever in a loop, accepting new TCP connections.

**3. Client connection accepted**
- When a client connects, the server accepts it and starts two goroutines:
- One goroutine runs `handleConnection(server, conn)` to manage that client.
- One goroutine runs `server.Broadcasts()` to process joins, leaves, and messages.

**4. Client join flow**
- `handleConnection` checks if the server already has 10 clients.
- If full, it sends `Server full. Maximum 10 clients allowed.` and closes the connection.
- If there is space, it calls `cmd.HandleClient(conn, server)`.
- `HandleClient` sends the ASCII banner (`utils.Banner`) to the client.
- The client types a name and presses Enter.
- The name is trimmed and stored in the `Client` struct.
- The client is sent into the `server.Join` channel.
- The server’s `Broadcasts` loop receives the join event, then:
  - Adds the client to the `Clients` map.
  - Sends the chat history to the new client.
  - Broadcasts a system message like `alice has joined our chat.` to everyone else.

**5. Client sends a message**
- Each client has a `ReadInput` loop running in a goroutine.
- It scans one line at a time from the TCP connection.
- Empty or whitespace‑only messages are ignored.
- Each valid message is turned into a `Message` object and sent to `server.Broadcast`.
- The server formats it as:
  ```
  [YYYY-MM-DD HH:MM:SS][name]: message
  ```
- The server stores the formatted line in history.
- The server broadcasts it to all clients **except** the sender.
- If you use the TUI client, it prints your own message locally so you still see it.

**6. Client leaves**
- If a client disconnects or closes the connection, `ReadInput` ends.
- The client is sent into `server.Leave`.
- The server broadcasts a system message like `alice has left our chat.`
- The client is removed from the `Clients` map.

**7. Changing a name**
- The current code does **not** support changing names after the first prompt.
- The name is read once in `cmd.HandleClient` and then stored in `Client.Name`.
- If you want `/nick newname` or similar behavior, you would need to do these steps:
1. Detect that command in `(*Client).ReadInput`.
2. Update the `Client.Name` safely (with a lock).
3. Update the `Clients` map key so it matches the new name.
4. Broadcast a system message like `oldname is now known as newname`.

**Project Structure and Key Functions**

`main.go`
- `main()`: Reads an optional port argument and starts the TCP server.

`server/server.go`
- `Start(port string)`: Starts the server with default output settings.
- `StartWithOptions(port string, opts Options)`: Starts the server with custom logging/output.
- `handleConnection(s *service.Server, conn net.Conn)`: Enforces max clients and hands the connection to the client handler.

`cmd/client.go`
- `HandleClient(c net.Conn, s *service.Server)`: Sends the banner, reads the user name, registers the client, and starts read/write loops.

`service/model.go`
- `DefaultPort`: The default port (`8989`).
- `Client`, `Server`, `Message`: Core data structures for the chat system.

`service/client.go`
- `(*Client).ReadInput(s *Server)`: Reads lines from the client connection and sends them into the broadcast channel. Removes the client on disconnect.
- `(*Client).WriteOutput()`: Writes messages from the server to the client connection.

`service/broadcast.go`
- `NewServer(maxConn int)`: Creates the server state (clients map, channels, history).
- `(*Server).Broadcasts()`: Main event loop. Handles join, leave, and chat messages.
- `(*Server).broadcastToOthers(...)`: Sends a message to all connected clients except the sender.
- `(*Server).addToHistory(...)`: Stores messages in memory.
- `formatUserMessage(...)`: Builds a timestamped user message string.
- `formatSystemMessage(...)`: Builds a timestamped system message string.

`utils/banner.go`
- `Banner`: The ASCII welcome banner shown to every new client.

`cmd/tui/main.go` (optional TUI client)
- `main()`: Parses flags, connects to the server, starts the UI.
- `parseAddr(...)`: Builds a host:port address from user input.
- `startLocalServer(...)`: Starts a server in the background for `--with-server` mode.
- `dialWithRetry(...)`: Re-tries connection until the local server is ready.
- `layout(...)`: Builds the chat and input views.
- `handleEnter(...)`: Sends name or chat messages when you press Enter.
- `readLoop()`: Reads messages from the server and prints them in the UI.

**How to Extend or Modify**
Here are common changes you can make:
- Change the default port: edit `service/model.go` (`DefaultPort`).
- Increase the client limit: edit `maxClients` in `server/server.go`.
- Customize the banner: edit `utils/banner.go`.
- Show your own messages on the server side: change `broadcastToOthers` to also send to the sender.
- Save chat history to disk: update `addToHistory` and load history on startup.
- Add commands (like `/nick` or `/list`): parse messages in `ReadInput` before broadcasting.
- Add TLS for secure connections: replace `net.Listen` and `net.Dial` with the TLS versions in `crypto/tls`.

**Running Tests**
```bash
go test ./...
```
