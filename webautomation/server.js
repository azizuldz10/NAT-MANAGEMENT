/**
 * ONT WiFi Extractor - Web UI Server
 * Express server untuk web interface
 */

const express = require('express');
const http = require('http');
const socketIo = require('socket.io');
const path = require('path');
const ONTLauncher = require('./ont-extractor-launcher.js');

const app = express();
const server = http.createServer(app);
const io = socketIo(server);

const PORT = process.env.PORT || 3000;

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use(express.static(path.join(__dirname, 'public')));

// Store active extractions
const activeExtractions = new Map();

// Socket.IO connection
io.on('connection', (socket) => {
    console.log('Client connected:', socket.id);

    socket.on('start-extraction', async (data) => {
        const { url, username, password, debug } = data;
        const extractionId = socket.id;

        console.log(`[${extractionId}] Starting extraction for: ${url}`);

        socket.emit('log', {
            type: 'info',
            message: `Starting extraction for ${url}...`
        });

        try {
            // Create launcher with custom logging
            const launcher = new ONTLauncher(
                url,
                username || null,
                password || null,
                debug || false
            );

            // Override log function to emit to socket
            const originalLog = launcher.log.bind(launcher);
            launcher.log = (message) => {
                originalLog(message);
                socket.emit('log', {
                    type: 'info',
                    message: message
                });
            };

            // Store active extraction
            activeExtractions.set(extractionId, {
                url,
                status: 'running',
                startTime: new Date()
            });

            // Emit progress update
            socket.emit('status', { status: 'detecting' });

            // Run extraction
            const result = await launcher.extract();

            if (result.success) {
                socket.emit('status', { status: 'success' });
                socket.emit('result', {
                    success: true,
                    model: result.model,
                    data: result.data,
                    credentials: result.credentials
                });

                console.log(`[${extractionId}] Extraction successful`);
            } else {
                socket.emit('status', { status: 'error' });
                socket.emit('result', {
                    success: false,
                    error: result.error,
                    model: result.model
                });

                console.log(`[${extractionId}] Extraction failed: ${result.error}`);
            }

        } catch (error) {
            socket.emit('status', { status: 'error' });
            socket.emit('result', {
                success: false,
                error: error.message
            });

            console.error(`[${extractionId}] Error:`, error.message);
        } finally {
            activeExtractions.delete(extractionId);
        }
    });

    socket.on('disconnect', () => {
        console.log('Client disconnected:', socket.id);
        activeExtractions.delete(socket.id);
    });
});

// REST API endpoints
app.get('/api/health', (req, res) => {
    res.json({
        status: 'ok',
        activeExtractions: activeExtractions.size,
        uptime: process.uptime()
    });
});

app.get('/api/credentials-template', (req, res) => {
    try {
        const templatePath = path.join(__dirname, 'ont-credentials-template.json');
        const fs = require('fs');
        const template = JSON.parse(fs.readFileSync(templatePath, 'utf8'));
        res.json(template);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Start server
server.listen(PORT, () => {
    console.log('\n' + '='.repeat(70));
    console.log('🌐 ONT WiFi Extractor - Web UI Server');
    console.log('='.repeat(70));
    console.log(`📡 Server running at: http://localhost:${PORT}`);
    console.log(`🔧 Environment: ${process.env.NODE_ENV || 'development'}`);
    console.log(`🚀 Ready to accept connections!`);
    console.log('='.repeat(70) + '\n');
});

// Graceful shutdown
process.on('SIGTERM', () => {
    console.log('SIGTERM signal received: closing HTTP server');
    server.close(() => {
        console.log('HTTP server closed');
        process.exit(0);
    });
});

module.exports = { app, server, io };
