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
                    connection.socket.write(JSON.stringify(data) + '\n');
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
        if (connection) return connection;

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
                    socketio: socket
                };

                socketToPipec.on('data', (data) => {
                    buffer += data.toString();
                    const lines = buffer.split('\n');
                    buffer = lines.pop();

                    for (const line of lines) {
                        if (line.trim()) {
                            try {
                                const response = JSON.parse(line);
                                if (socket.connected) {
                                    socket.emit('pipec-response', response);
                                }
                            } catch (err) {
                                console.error(`Failed to parse PIPEC response for ${clientId}:`, err);
                            }
                        }
                    }
                });

                socketToPipec.on('error', (err) => {
                    console.error(`PIPEC connection error for ${clientId}:`, err);
                    this.connections.delete(clientId);
                });

                socketToPipec.on('close', () => {
                    console.log(`PIPEC connection closed for client ${clientId}`);
                    this.connections.delete(clientId);
                });

                this.connections.set(clientId, conn);
                resolve(conn);
            });
        });
    }

    closeConnection(clientId) {
        const connection = this.connections.get(clientId);
        if (connection) {
            connection.socket.destroy();
            this.connections.delete(clientId);
            console.log(`Closed PIPEC connection for client ${clientId}`);
        }
    }

    sendError(socket, message) {
        if (socket.connected) {
            socket.emit('error', { status: 'ERROR', message });
        }
    }

    startCleanupTimer() {
        this.cleanupInterval = setInterval(() => {
            const now = Date.now();
            const idleTimeout = 30000;
            for (const [clientId, conn] of this.connections.entries()) {
                if (now - conn.lastActivity > idleTimeout) {
                    console.log(`Cleaning up idle connection for client ${clientId}`);
                    this.closeConnection(clientId);
                }
            }
        }, 10000);
    }

    stop() {
        if (this.cleanupInterval) clearInterval(this.cleanupInterval);
        for (const clientId of this.connections.keys()) this.closeConnection(clientId);
        console.log('Socket.IO proxy server stopped');
    }
}

module.exports = PipecProxy;
