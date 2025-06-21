require('dotenv').config();
const PipecProxy = require('./PipecProxy');
const http = require('http');
const { Server } = require('socket.io');

const proxy = new PipecProxy({
    pipecHost: process.env.PIPEC_HOST || 'localhost',
    pipecPort: process.env.PIPEC_PORT || 1143
});

const httpServer = http.createServer();

const io = new Server(httpServer, {
    cors: { origin: '*' },
});

proxy.start(io);

httpServer.listen(process.env.PROXY_PORT || 96, () => {
    console.log(`Socket.IO proxy server listening on port ${process.env.PROXY_PORT || 96}`);
});

function startHealthCheck() {
    const healthPort = process.env.HEALTH_PORT || 3001;
    const server = http.createServer((req, res) => {
        if (req.url === '/health') {
            res.writeHead(200, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify({
                status: 'ok',
                timestamp: new Date().toISOString(),
                connections: proxy.connections.size
            }));
        } else {
            res.writeHead(404);
            res.end('Not Found');
        }
    });

    server.listen(healthPort, () => {
        console.log(`Health check server listening on port ${healthPort}`);
    });

    return server;
}

const healthServer = startHealthCheck();

function gracefulShutdown() {
    console.log('Shutting down gracefully...');
    proxy.stop();
    io.close();
    healthServer.close();
    process.exit(0);
}

process.on('SIGTERM', gracefulShutdown);
process.on('SIGINT', gracefulShutdown);
process.on('uncaughtException', (err) => {
    console.error('Uncaught Exception:', err);
    proxy.stop();
    process.exit(1);
});
process.on('unhandledRejection', (reason, promise) => {
    console.error('Unhandled Rejection at:', promise, 'reason:', reason);
    proxy.stop();
    process.exit(1);
});
