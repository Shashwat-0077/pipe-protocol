// File: proxy/PipecProxy.js
const WebSocket = require('ws');
const net = require('net');
const { v4: uuidv4 } = require('uuid');

class PipecProxy {
    constructor({ pipecHost = 'localhost', pipecPort = 1143, port = 96 } = {}) {
        this.port = port;
        this.pipecHost = pipecHost;
        this.pipecPort = pipecPort;
        this.connections = new Map();
        this.cleanupInterval = null;
        console.log(`Proxy will connect to PIPEC at ${this.pipecHost}:${this.pipecPort}`);
    }

    start() {
        const wss = new WebSocket.Server({ port: this.port, perMessageDeflate: false });
        console.log(`WebSocket proxy server listening on port ${this.port}`);

        wss.on('connection', (ws, req) => {
            const clientId = uuidv4();
            const clientIP = req.socket.remoteAddress;
            console.log(`New WebSocket client connected: ${clientId} from ${clientIP}`);

            ws.on('message', async (message) => {
                try {
                    const data = JSON.parse(message.toString());
                    await this.handleMessage(ws, clientId, data);
                } catch (error) {
                    console.error(`Error processing message from ${clientId}:`, error);
                    this.sendError(ws, 'Invalid JSON format');
                }
            });

            ws.on('close', () => {
                console.log(`WebSocket client disconnected: ${clientId}`);
                this.closeConnection(clientId);
            });

            ws.on('error', (error) => {
                console.error(`WebSocket error for client ${clientId}:`, error);
                this.closeConnection(clientId);
            });
        });

        this.startCleanupTimer();
        return wss;
    }

    async handleMessage(ws, clientId, data) {
        try {
            let connection = this.connections.get(clientId);
            if (!connection) {
                connection = await this.createPipecConnection(clientId);
                if (!connection) return this.sendError(ws, 'Failed to connect to PIPEC server');
            }
            connection.lastActivity = Date.now();
            connection.socket.write(JSON.stringify(data) + '\n');
            if (!connection.responseHandler) {
                connection.responseHandler = (response) => {
                    if (ws.readyState === WebSocket.OPEN) {
                        ws.send(JSON.stringify(response));
                    }
                };
            }
        } catch (error) {
            console.error(`Error handling message for ${clientId}:`, error);
            this.sendError(ws, 'Internal server error');
        }
    }

    async createPipecConnection(clientId) {
        return new Promise((resolve, reject) => {
            const socket = new net.Socket();
            let buffer = '';
            let isConnected = false;

            const timeout = setTimeout(() => {
                if (!isConnected) {
                    socket.destroy();
                    reject(new Error('Connection timeout'));
                }
            }, 5000);

            socket.connect(this.pipecPort, this.pipecHost, () => {
                clearTimeout(timeout);
                isConnected = true;
                console.log(`Connected to PIPEC server for client ${clientId}`);

                const connection = {
                    socket,
                    lastActivity: Date.now(),
                    responseHandler: null
                };

                this.connections.set(clientId, connection);
                resolve(connection);
            });

            socket.on('data', (data) => {
                buffer += data.toString();
                const lines = buffer.split('\n');
                buffer = lines.pop();
                for (const line of lines) {
                    if (line.trim()) {
                        try {
                            const response = JSON.parse(line);
                            const conn = this.connections.get(clientId);
                            if (conn?.responseHandler) {
                                conn.responseHandler(response);
                            }
                        } catch (error) {
                            console.error(`Error parsing PIPEC response for ${clientId}:`, error);
                        }
                    }
                }
            });

            socket.on('error', (error) => {
                clearTimeout(timeout);
                console.error(`PIPEC connection error for client ${clientId}:`, error);
                this.connections.delete(clientId);
                if (!isConnected) reject(error);
            });

            socket.on('close', () => {
                clearTimeout(timeout);
                console.log(`PIPEC connection closed for client ${clientId}`);
                this.connections.delete(clientId);
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

    sendError(ws, message) {
        if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ status: 'ERROR', message }));
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
        console.log('Proxy server stopped');
    }
}

module.exports = PipecProxy;
