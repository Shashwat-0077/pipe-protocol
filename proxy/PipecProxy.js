const net = require('net');
const { v4: uuidv4 } = require('uuid');

class PipecProxy {
    constructor({ pipecHost = 'localhost', pipecPort = 1143 } = {}) {
        this.pipecHost = pipecHost;
        this.pipecPort = pipecPort;
        this.connections = new Map();
        this.cleanupInterval = null;
        console.log(`Proxy will connect to PIPEC at ${this.pipecHost}:${this.pipecPort}`);
    }

    start(io) {
        io.on('connection', (socket) => {
            const clientId = uuidv4();
            const clientIP = socket.handshake.address;
            console.log(`Socket.IO client connected: ${clientId} from ${clientIP}`);

            socket.on('pipec-message', async (data) => {
                try {
                    const connection = await this.getOrCreateConnection(clientId, socket);
                    connection.lastActivity = Date.now();

                    const enrichedData = {
                        ...data,
                        connection_type: 'client'
                    };

                    console.log(`Handling message for ${clientId}:`, enrichedData);

                    connection.socket.write(JSON.stringify(enrichedData) + '\n');
                } catch (error) {
                    console.error(`Error handling message for ${clientId}:`, error);
                    this.sendError(socket, 'Internal server error');
                }
            });

            socket.on('disconnect', () => {
                console.log(`Socket.IO client disconnected: ${clientId}`);
                this.closeConnection(clientId);
            });

            socket.on('error', (error) => {
                console.error(`Socket.IO error for client ${clientId}:`, error);
                this.closeConnection(clientId);
            });
        });

        this.startCleanupTimer();
    }

    async getOrCreateConnection(clientId, socket) {
        let connection = this.connections.get(clientId);
        if (connection) {
            connection.lastActivity = Date.now();
            return connection;
        }

        return new Promise((resolve, reject) => {
            const socketToPipec = new net.Socket();
            let buffer = '';
            let isConnected = false;

            const timeout = setTimeout(() => {
                if (!isConnected) {
                    socketToPipec.destroy();
                    reject(new Error('Connection timeout'));
                }
            }, 5000);

            socketToPipec.connect(this.pipecPort, this.pipecHost, () => {
                clearTimeout(timeout);
                isConnected = true;
                console.log(`Connected to PIPEC server for client ${clientId}`);

                const conn = {
                    socket: socketToPipec,
                    lastActivity: Date.now(),
                    socketio: socket,
                    clientId
                };

                socketToPipec.on('data', (data) => {
                    buffer += data.toString();
                    const lines = buffer.split('\n');
                    buffer = lines.pop();

                    for (const line of lines) {
                        if (line.trim()) {
                            try {
                                const response = JSON.parse(line);
                                const currentConn = this.connections.get(clientId);
                                if (currentConn && currentConn.socketio?.connected) {
                                    currentConn.lastActivity = Date.now();
                                    currentConn.socketio.emit('pipec-response', response);
                                }
                            } catch (err) {
                                console.error(`Failed to parse PIPEC response for ${clientId}:`, err);
                            }
                        }
                    }
                });

                socketToPipec.on('error', (err) => {
                    console.error(`PIPEC connection error for ${clientId}:`, err);
                    this.closeConnection(clientId);
                });

                socketToPipec.on('close', () => {
                    console.log(`PIPEC connection closed for client ${clientId}`);
                    this.closeConnection(clientId);
                });

                this.connections.set(clientId, conn);
                resolve(conn);
            });

            socketToPipec.on('error', (err) => {
                clearTimeout(timeout);
                if (!isConnected) {
                    reject(err);
                }
            });
        });
    }

    closeConnection(clientId) {
        const connection = this.connections.get(clientId);
        if (connection) {
            try {
                if (connection.socket && !connection.socket.destroyed) {
                    connection.socket.destroy();
                }
            } catch (err) {
                console.error(`Error destroying socket for ${clientId}:`, err);
            }
            this.connections.delete(clientId);
            console.log(`Closed PIPEC connection for client ${clientId}`);
        }
    }

    sendError(socket, message) {
        if (socket && socket.connected) {
            socket.emit('error', { status: 'ERROR', message });
        }
    }

    startCleanupTimer() {
        this.cleanupInterval = setInterval(() => {
            const now = Date.now();
            const idleTimeout = 30 * 1000;
            const connectionsToClose = [];

            for (const [clientId, conn] of this.connections.entries()) {
                const timeSinceLastActivity = now - conn.lastActivity;
                if (timeSinceLastActivity > idleTimeout || !conn.socketio?.connected) {
                    connectionsToClose.push(clientId);
                    console.log(`Marking connection ${clientId} for cleanup - idle for ${Math.floor(timeSinceLastActivity / 1000)}s, socket.io connected: ${conn.socketio?.connected || false}`);
                }
            }

            for (const clientId of connectionsToClose) {
                this.closeConnection(clientId);
            }

            if (Math.floor(now / 1000) % 300 === 0) {
                console.log(`Active connections: ${this.connections.size}`);
            }
        }, 10 * 1000);
    }

    stop() {
        if (this.cleanupInterval) {
            clearInterval(this.cleanupInterval);
            this.cleanupInterval = null;
        }

        for (const clientId of this.connections.keys()) {
            this.closeConnection(clientId);
        }

        console.log('Socket.IO proxy server stopped');
    }
}

module.exports = PipecProxy;
