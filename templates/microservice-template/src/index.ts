// @ts-nocheck
import 'dotenv/config';
import { AddressInfo } from 'net';
import { Server } from 'http';
import app from './app.js';
import env from './config/env.js';
import { markShuttingDown, isShuttingDown } from '@/utils/server/shutdown.js';
import { logger } from '@zerodayz7/common';
import { Socket } from 'net';

// Server instance with proper typing
let server: Server;

// Configuration constants
const SERVER_CONFIG = {
  SHUTDOWN_TIMEOUT: 30_000,
  KEEP_ALIVE_TIMEOUT: 61_000,
  HEADERS_TIMEOUT: 65_000,
  MAX_CONNECTIONS: env.MAX_CONNECTIONS,
  REQUEST_TIMEOUT: env.REQUEST_TIMEOUT,
} as const;

// Log server information
function logServerInfo(server: Server): void {
  const address = server.address();
  if (address && typeof address === 'object') {
    logger.info(`🚀 Server running on port ${(address as AddressInfo).port}`);
  } else {
    logger.info(`🚀 Server running (port unknown)`);
  }
  logger.info(`Mode: ${env.NODE_ENV}`);
  logger.info(`Version: ${env.APP_VERSION}`);
  logger.info(`PID: ${process.pid}`);
  logger.info(`Maximum connections: ${SERVER_CONFIG.MAX_CONNECTIONS}`);
  logger.info(`Keep-alive timeout: ${SERVER_CONFIG.KEEP_ALIVE_TIMEOUT}ms`);
}

// Server configuration
function setupServerConfiguration(server: Server): void {
  // Performance optimizations
  server.keepAliveTimeout = SERVER_CONFIG.KEEP_ALIVE_TIMEOUT;
  server.headersTimeout = SERVER_CONFIG.HEADERS_TIMEOUT;
  server.requestTimeout = SERVER_CONFIG.REQUEST_TIMEOUT;
  server.maxConnections = SERVER_CONFIG.MAX_CONNECTIONS;

  // Error handling
  server.on('error', (err: NodeJS.ErrnoException) => {
    if (err.code === 'EADDRINUSE') {
      logger.error(`Port ${env.PORT} is already in use.`);
    } else if (err.code === 'EACCES') {
      logger.error(`Permission denied to bind to port ${env.PORT}`);
    } else {
      logger.error('Server error:', { error: err.message, code: err.code, stack: err.stack });
    }
    process.exit(1);
  });

  server.on('clientError', (error: Error, socket: Socket) => {
    logger.warn(`Client error: ${error.message}`, {
      remoteAddress: socket.remoteAddress,
      remotePort: socket.remotePort,
    });

    if (socket.writable && !socket.destroyed) {
      socket.end('HTTP/1.1 400 Bad Request\r\n\r\n');
    }
  });

  // Connection monitoring
  server.on('connection', (socket) => {
    socket.on('error', (err) => {
      logger.debug('Socket error:', err.message);
    });
  });
}

// Initialize services
async function initializeServices(): Promise<void> {
  try {
    logger.info('🔧 Initializing services...');
  } catch (error) {
    logger.error('Failed to initialize services:', error);
    throw error;
  }
}

// Start server
async function startServer(): Promise<Server> {
  try {
    await initializeServices();
    return new Promise((resolve, reject) => {
      const serverInstance = app.listen(env.PORT, (err?: Error) => {
        if (err) {
          reject(err);
          return;
        }
        server = serverInstance;
        // if (server) {
          setupServerConfiguration(server);
          logServerInfo(server);
          resolve(server);
        // }
      });
    });
  } catch (error) {
    logger.error('Failed to start server:', error);
    throw error;
  }
}

// Graceful shutdown
async function gracefulShutdown(signal?: string): Promise<void> {
  if (isShuttingDown()) return;
  markShuttingDown();

  logger.info(`🛑 Received ${signal || 'shutdown signal'}, starting graceful shutdown...`);

  const timer = setTimeout(() => {
    logger.error('Forced shutdown due to timeout');
    process.exit(1);
  }, SERVER_CONFIG.SHUTDOWN_TIMEOUT);

  try {
    if (server) {
      await new Promise<void>((resolve, reject) => server!.close((err) => (err ? reject(err) : resolve())));
      logger.info('✅ Server closed');
    }

    logger.info('✅ Redis connections closed');
    clearTimeout(timer);
    logger.info('👋 Graceful shutdown complete. Goodbye!');
    process.exit(0);
  } catch (error) {
    logger.error('Error during shutdown:', error);
    clearTimeout(timer);
    process.exit(1);
  }
}

// Signal handlers
['SIGINT', 'SIGTERM', 'SIGUSR2'].forEach((sig) => {
  process.on(sig, () => gracefulShutdown(sig));
});

// Error handling
process.on('uncaughtException', (err) => {
  logger.error('Uncaught Exception:', {
    error: err.message,
    stack: err.stack,
    pid: process.pid,
  });
  if (env.NODE_ENV === 'production') {
    process.exit(1);
  }
});

process.on('unhandledRejection', (reason, promise) => {
  logger.error('Unhandled Rejection:', {
    reason: reason,
    promise: promise,
    pid: process.pid,
  });
  if (env.NODE_ENV === 'production') {
    process.exit(1);
  }
});

// Application startup
async function bootstrap(): Promise<void> {
  try {
    await startServer();
    logger.info('🎉 Application started successfully');
  } catch (error) {
    logger.error('Failed to start application:', error);
    process.exit(1);
  }
}

// Start application
bootstrap().catch((error) => {
  logger.error('Bootstrap failed:', error);
  process.exit(1);
});
