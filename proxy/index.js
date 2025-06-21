// File: proxy/index.js
require('dotenv').config();
const PipecProxy = require('./PipecProxy');
const http = require('http');

const proxy = new PipecProxy({
    port: process.env.PROXY_PORT || 96,
    pipecHost: process.env.PIPEC_HOST || 'localhost',
    pipecPort: process.env.PIPEC_PORT || 1143
});

const wss = proxy.start();

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

process.on('SIGTERM', () => {
    console.log('Received SIGTERM, shutting down gracefully...');
    proxy.stop();
    wss.close();
    healthServer.close();
    process.exit(0);
});

process.on('SIGINT', () => {
    console.log('Received SIGINT, shutting down gracefully...');
    proxy.stop();
    wss.close();
    healthServer.close();
    process.exit(0);
});

process.on('uncaughtException', (error) => {
    console.error('Uncaught Exception:', error);
    proxy.stop();
    process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
    console.error('Unhandled Rejection at:', promise, 'reason:', reason);
    proxy.stop();
    process.exit(1);
});
