require("dotenv").config();
const express = require("express");
const httpProxy = require("http-proxy");
const { Pool } = require("pg");
const helmet = require("helmet");
const compression = require("compression");
const cors = require("cors");
const rateLimit = require("express-rate-limit");
const winston = require("winston");
const morgan = require("morgan"); // For HTTP request logging
const path = require("path");

// Initialize app and logger
const app = express();
const logger = winston.createLogger({
  level: process.env.LOG_LEVEL || "info",
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.json()
  ),
  transports: [
    new winston.transports.Console({ format: winston.format.simple() }),
    new winston.transports.File({ filename: "app.log" }),
  ],
});

const PORT = process.env.PORT || 8000;
const S3_BASE_PATH = process.env.S3_BASE_PATH;

const requiredEnvVars = [
  "PORT",
  "S3_BASE_PATH",
  "DB_PORT",
  "DB_USER",
  "DB_DATABASE",
  "DB_PASSWORD",
  "DB_HOST",
];

requiredEnvVars.forEach((varName) => {
  if (!process.env[varName] || process.env[varName].trim() === "") {
    console.error(`ERROR: Missing or empty environment variable: ${varName}`);
    process.exit(1);
  }
});

// Database connection
const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_PORT,
  user: process.env.DB_USER,
  database: process.env.DB_DATABASE,
  password: process.env.DB_PASSWORD,
  ssl: { rejectUnauthorized: false },
});

// Proxy server setup
const proxy = httpProxy.createProxy();

// Middleware
app.use(helmet()); // Security headers
app.use(compression()); // GZIP compression
app.use(cors({ origin: process.env.CORS_ORIGIN || "*" })); // CORS, configure properly
app.use(rateLimit({ windowMs: 15 * 60 * 1000, max: 100 })); // Rate limiting
app.use(
  morgan("combined", {
    stream: { write: (message) => logger.info(message.trim()) },
  })
); // HTTP request logging

// Route to handle proxying and deployment resolution
app.use(async (req, res) => {
  try {
    const hostname = req.hostname;
    const subdomain = hostname.split(".")[0];

    // Query projects table
    const projectQuery = `SELECT id FROM projects WHERE subdomain = $1 LIMIT 1;`;
    const projectResult = await pool.query(projectQuery, [subdomain]);

    if (projectResult.rows.length === 0) {
      logger.warn(`Subdomain not found: ${subdomain}`);
      return res.status(404).send("Subdomain not found");
    }
    const projectId = projectResult.rows[0].id;

    // Query deployments table
    const deploymentQuery = `
      SELECT id FROM deployments 
      WHERE project_id = $1 AND status = 'READY' 
      ORDER BY created_at DESC 
      LIMIT 1;
    `;
    const deploymentResult = await pool.query(deploymentQuery, [projectId]);

    if (deploymentResult.rows.length === 0) {
      logger.warn(`No deployment found for projectId: ${projectId}`);
      return res.status(404).send("No deployment found with status 'READY'");
    }
    const deploymentId = deploymentResult.rows[0].id;

    // Proxy request to deployment target
    const resolvesTo = `${S3_BASE_PATH}/${deploymentId}`;
    logger.info(`Proxying request to: ${resolvesTo}`);
    return proxy.web(req, res, { target: resolvesTo, changeOrigin: true });
  } catch (err) {
    logger.error("Error handling request", { error: err });
    res.status(500).send("Internal Server Error");
  }
});

// Proxy modifications
proxy.on("proxyReq", (proxyReq, req, res) => {
  const url = req.url;
  if (url === "/") proxyReq.path += "index.html";
});

proxy.on("error", (err, req, res) => {
  logger.error("Proxy error", { error: err });
  res.status(502).send("Bad Gateway");
});

// Start the server
const server = app.listen(PORT, () => {
  logger.info(`Server running on port ${PORT}`);
});

// Graceful shutdown
const shutdown = async (signal) => {
  logger.info(`Received ${signal}, shutting down server...`);
  server.close(() => {
    logger.info("HTTP server closed.");
    pool.end(() => {
      logger.info("Database pool closed.");
      process.exit(0);
    });
  });

  // If the server takes too long to close, force shutdown
  setTimeout(() => {
    logger.error("Forcing server shutdown due to timeout...");
    process.exit(1);
  }, 10000); // 10 seconds timeout
};

process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);
